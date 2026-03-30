package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"todo-api/models/db"
	userModel "todo-api/models/user"
	"todo-api/services/auth"
	"todo-api/utils"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
)

type AuthController struct {
	beego.Controller
}

const (
	accessTokenCookieName  = "accesstoken"
	refreshTokenCookieName = "refreshtoken"
)

func (c *AuthController) authService() *auth.AuthService {
	repo := &userModel.UserRepository{DB: db.DB}
	secret, err := beego.AppConfig.String("secret::JWT_SECRET")

	if err != nil {
		log.Fatal("Failed to read secret key from config")
		panic(fmt.Sprintf("Failed to read secret key from config: %v", err))
	}
	if secret == "" {
		log.Fatal("Secret key is not set in app.conf")
		panic("Secret key is not set in app.conf")
	}

	return auth.NewAuthService(repo, secret)
}

func (c *AuthController) handleAuthError(action string, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, auth.ErrInvalidAuthInput),
		errors.Is(err, auth.ErrInvalidEmailFormat):
		utils.RespondWithError(c.Ctx, 400, err.Error())
	case errors.Is(err, auth.ErrUserAlreadyExists):
		utils.RespondWithError(c.Ctx, 409, err.Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		utils.RespondWithError(c.Ctx, 401, err.Error())
	default:
		log.Printf("auth %s failed: %v", action, err)
		utils.RespondWithError(c.Ctx, 500, fmt.Sprintf("failed to %s", action))
	}
}

func decodeAuthJSONBody(body io.Reader, dst any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("invalid request body: request body must contain a single JSON object")
	}

	return nil
}

func (c *AuthController) Register() {
	var user userModel.User
	if err := decodeAuthJSONBody(c.Ctx.Request.Body, &user); err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}
	authService := c.authService()

	err := authService.Register(&user)
	if err != nil {
		c.handleAuthError("register user", err)
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]string{"message": "user created"}
	c.ServeJSON()
}

func (c *AuthController) Login() {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := decodeAuthJSONBody(c.Ctx.Request.Body, &input); err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}
	authService := c.authService()

	accessToken, refreshToken, err := authService.Login(input.Email, input.Password)
	if err != nil {
		c.handleAuthError("login", err)
		return
	}

	isSecureCookie := strings.ToLower(strings.TrimSpace(beego.BConfig.RunMode)) == "prod"

	http.SetCookie(c.Ctx.ResponseWriter, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessToken,
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
		Secure:   isSecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Ctx.ResponseWriter, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   isSecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	c.Data["json"] = map[string]string{"accesstoken": accessToken}
	c.ServeJSON()
}

func (c *AuthController) Logout() {
	isSecureCookie := strings.ToLower(strings.TrimSpace(beego.BConfig.RunMode)) == "prod"

	expiredCookie := func(name string) *http.Cookie {
		return &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   isSecureCookie,
			SameSite: http.SameSiteLaxMode,
		}
	}

	http.SetCookie(c.Ctx.ResponseWriter, expiredCookie(accessTokenCookieName))
	http.SetCookie(c.Ctx.ResponseWriter, expiredCookie(refreshTokenCookieName))

	c.Data["json"] = map[string]string{"message": "logged out"}
	c.ServeJSON()
}

func extractIntClaim(claims jwt.MapClaims, key string) (int, error) {
	value, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("missing %s in token", key)
	}

	switch v := value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("invalid %s in token", key)
	}
}

func extractStringClaim(claims jwt.MapClaims, key string) (string, bool) {
	value, ok := claims[key]
	if !ok {
		return "", false
	}

	strValue, ok := value.(string)
	if !ok {
		return "", false
	}

	return strValue, true
}

func (c *AuthController) Me() {
	accessCookie, err := c.Ctx.Request.Cookie(accessTokenCookieName)
	if err != nil || strings.TrimSpace(accessCookie.Value) == "" {
		utils.RespondWithError(c.Ctx, 401, "access token cookie is required")
		return
	}

	secret := strings.TrimSpace(c.authService().JwtSecret)
	tokenString := strings.TrimSpace(accessCookie.Value)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		utils.RespondWithError(c.Ctx, 401, "invalid or expired token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.RespondWithError(c.Ctx, 401, "invalid token claims")
		return
	}

	expiry, err := claims.GetExpirationTime()
	if err != nil || expiry == nil || expiry.Time.Before(time.Now()) {
		utils.RespondWithError(c.Ctx, 401, "token is expired")
		return
	}

	userID, err := extractIntClaim(claims, "user_id")
	if err != nil {
		utils.RespondWithError(c.Ctx, 401, err.Error())
		return
	}

	username, ok := extractStringClaim(claims, "username")
	if !ok || strings.TrimSpace(username) == "" {
		utils.RespondWithError(c.Ctx, 401, "missing username in token")
		return
	}

	response := map[string]any{
		"id":       userID,
		"username": username,
	}

	c.Data["json"] = response
	c.ServeJSON()
}
