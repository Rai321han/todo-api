package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"todo-api/models/db"
	userModel "todo-api/models/user"
	"todo-api/service/auth"
	"todo-api/utils"

	beego "github.com/beego/beego/v2/server/web"
)


type AuthController struct {
	beego.Controller
}

func (c *AuthController) authService() *auth.AuthService {
	repo := &userModel.UserRepository{DB: db.DB}
	// get secret form app.conf
	secret, err := beego.AppConfig.String("jwt::JWT_SECRET")
	
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

func decodeAuthJSONBody(body io.Reader, dst interface{}) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	var extra interface{}
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

	token, err := authService.Login(input.Email, input.Password)
	if err != nil {
		c.handleAuthError("login", err)
		return
	}

	c.Data["json"] = map[string]string{"token": token}
	c.ServeJSON()
}
