package middlewares

import (
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware is a Beego middleware function that checks for the presence of a valid JWT token in the Authorization header of incoming HTTP requests.
// It expects the token to be in the format "Bearer <token>".
// If the token is missing, invalid, or expired, it responds with a 401 Unauthorized status and an appropriate error message.
// If the token is valid, the middleware allows the request to proceed to the next handler in the chain.
func AuthMiddleware(ctx *context.Context) {
	authHeader := ctx.Input.Header("Authorization")

	if authHeader == "" {
		ctx.Output.SetStatus(401)
		ctx.Output.Body([]byte("Unauthorized: No token provided"))
		return
	}

	// Expect : Bearer <token>
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		ctx.Output.SetStatus(401)
		ctx.Output.Body([]byte("Unauthorized: Invalid token format"))
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	secret, _ := beego.AppConfig.String("secret::JWT_SECRET")

	hmacSampleSecret := []byte(secret)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {

	// hmacSampleSecret is a []byte containing the secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))


	if err != nil || !token.Valid {
		ctx.Output.SetStatus(401)
		ctx.Output.JSON(map[string]string{
			"error": "invalid or expired token",
		}, false, false)
		ctx.Abort(401, "Unauthorized")
		return
	}

	// Extract user info from token
	// jwt.MapClaims is a map[string]interface{}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		ctx.Input.SetData("user_id", int(claims["user_id"].(float64)))
		ctx.Input.SetData("username", claims["username"].(string))
	}
}