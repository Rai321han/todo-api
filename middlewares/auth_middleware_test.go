package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
	. "github.com/smartystreets/goconvey/convey"
)

// newTestContext creates a new Beego context with a test HTTP request and response recorder.
// It accepts an optional authHeader string to set the Authorization header in the request.
func newTestContext(t *testing.T, authHeader string) (*context.Context, *httptest.ResponseRecorder) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/v1/api/todos/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	res := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(res, req)

	return ctx, res
}

func jwtSecretForTest() string {
	secret, _ := beego.AppConfig.String("secret::JWT_SECRET")
	return secret
}

func makeToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign jwt token: %v", err)
	}

	return tokenString
}


// runAuthMiddlewareSafely executes the AuthMiddleware and recovers from any panic that occurs during its execution.
// It returns a boolean indicating whether a panic occurred.
// This is useful for testing scenarios where the middleware might panic due to invalid input or other issues, allowing the test to assert on the expected behavior without crashing the test suite.
func runAuthMiddlewareSafely(ctx *context.Context) (panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()

	AuthMiddleware(ctx)
	return false
}

func TestAuthMiddlewareUnauthorizedCases(t *testing.T) {
	Convey("AuthMiddleware unauthorized flows", t, func() {
		secret := jwtSecretForTest()

		tests := []struct {
			name            string
			authHeader      string
			wantStatus      int
			wantBodyContain string
			wantPanic       bool
		}{
			{
				name:            "missing authorization header",
				authHeader:      "",
				wantStatus:      http.StatusUnauthorized,
				wantBodyContain: "No token provided",
				wantPanic:       false,
			},
			{
				name:            "invalid header format",
				authHeader:      "Token abc",
				wantStatus:      http.StatusUnauthorized,
				wantBodyContain: "Invalid token format",
				wantPanic:       false,
			},
			{
				name:            "invalid jwt token",
				authHeader:      "Bearer " + makeToken(t, secret+"-wrong", jwt.MapClaims{"user_id": 7, "username": "alice", "exp": time.Now().Add(time.Hour).Unix()}),
				wantStatus:      http.StatusUnauthorized,
				wantBodyContain: "invalid or expired token",
				wantPanic:       true,
			},
		}

		for _, tc := range tests {
			tc := tc
			Convey(tc.name, func() {
				ctx, res := newTestContext(t, tc.authHeader)

				panicked := runAuthMiddlewareSafely(ctx)

				So(res.Code, ShouldEqual, tc.wantStatus)
				So(strings.Contains(res.Body.String(), tc.wantBodyContain), ShouldBeTrue)
				So(panicked, ShouldEqual, tc.wantPanic)
			})
		}
	})
}

func TestAuthMiddlewareValidTokenSetsContextData(t *testing.T) {
	Convey("AuthMiddleware valid token", t, func() {
		secret := jwtSecretForTest()
		tokenString := makeToken(t, secret, jwt.MapClaims{
			"user_id":  42,
			"username": "john",
			"exp":      time.Now().Add(time.Hour).Unix(),
		})

		ctx, res := newTestContext(t, "Bearer "+tokenString)

		AuthMiddleware(ctx)

		So(res.Code, ShouldNotEqual, http.StatusUnauthorized)

		userIDVal := ctx.Input.GetData("user_id")
		usernameVal := ctx.Input.GetData("username")

		So(userIDVal, ShouldNotBeNil)
		So(usernameVal, ShouldNotBeNil)
		So(userIDVal, ShouldHaveSameTypeAs, 0)
		So(usernameVal, ShouldHaveSameTypeAs, "")
		So(userIDVal.(int), ShouldEqual, 42)
		So(usernameVal.(string), ShouldEqual, "john")
	})
}

func TestAuthMiddlewareInvalidTokenReturnsJSONError(t *testing.T) {
	Convey("AuthMiddleware invalid token returns JSON payload", t, func() {
		ctx, res := newTestContext(t, "Bearer malformed.token.value")

		panicked := runAuthMiddlewareSafely(ctx)

		So(res.Code, ShouldEqual, http.StatusUnauthorized)
		So(panicked, ShouldBeTrue)

		var payload map[string]string
		err := json.Unmarshal(res.Body.Bytes(), &payload)
		So(err, ShouldBeNil)
		So(payload["error"], ShouldEqual, "invalid or expired token")
	})
}

