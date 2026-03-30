package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/smartystreets/goconvey/convey"

	userModel "todo-api/models/user"
)

type fakeUserRepo struct {
	getUserByEmailFunc func(email string) (*userModel.User, error)
	createFunc         func(user *userModel.User) error
}

func (f *fakeUserRepo) GetUserByEmail(email string) (*userModel.User, error) {
	if f.getUserByEmailFunc != nil {
		return f.getUserByEmailFunc(email)
	}
	return &userModel.User{}, userModel.ErrUserNotFound
}

func (f *fakeUserRepo) Create(user *userModel.User) error {
	if f.createFunc != nil {
		return f.createFunc(user)
	}
	return nil
}

const (
	fakeValidEmail = "user@example.com"
	fakeBadEmail   = "bad-email"
	fakeUsername   = "testuser"

	plainTextPassword = "plain-pass"
	wrongPassword     = "wrong-pass"
	jwtSecret         = "jwt_secret"
)

func TestIsValidEmail(t *testing.T) {
	Convey("isValidEmail should validate format", t, func() {
		So(isValidEmail(fakeValidEmail), ShouldBeTrue)
		So(isValidEmail(fakeBadEmail), ShouldBeFalse)
	})
}

func TestPasswordHelpers(t *testing.T) {
	Convey("HashPassword and CheckPasswordHash should work together", t, func() {
		hash, err := HashPassword("secret123")

		So(err, ShouldBeNil)
		So(hash, ShouldNotEqual, "secret123")
		So(CheckPasswordHash("secret123", hash), ShouldBeTrue)
		So(CheckPasswordHash("wrong", hash), ShouldBeFalse)
	})
}

func TestAuthServiceRegister(t *testing.T) {
	Convey("Register should reject invalid email", t, func() {
		svc := NewAuthService(&fakeUserRepo{}, jwtSecret)

		err := svc.Register(&userModel.User{Email: fakeBadEmail, Password: plainTextPassword})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidEmailFormat), ShouldBeTrue)
	})

	Convey("Register should reject missing username", t, func() {
		svc := NewAuthService(&fakeUserRepo{}, jwtSecret)

		err := svc.Register(&userModel.User{Email: fakeValidEmail, Password: plainTextPassword})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidAuthInput), ShouldBeTrue)
	})

	Convey("Register should reject missing password", t, func() {
		svc := NewAuthService(&fakeUserRepo{}, jwtSecret)

		err := svc.Register(&userModel.User{Email: fakeValidEmail, Username: fakeUsername})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidAuthInput), ShouldBeTrue)
	})

	Convey("Register should reject existing user", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				So(email, ShouldEqual, fakeValidEmail)
				return &userModel.User{ID: 1, Email: email}, nil
			},
		}
		svc := NewAuthService(repo, jwtSecret)

		err := svc.Register(&userModel.User{Username: fakeUsername, Email: fakeValidEmail, Password: plainTextPassword})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrUserAlreadyExists), ShouldBeTrue)
	})

	Convey("Register should hash password and create user", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{}, userModel.ErrUserNotFound
			},
			createFunc: func(user *userModel.User) error {
				So(user.Email, ShouldEqual, fakeValidEmail)
				So(user.Password, ShouldNotEqual, plainTextPassword)
				So(CheckPasswordHash(plainTextPassword, user.Password), ShouldBeTrue)
				return nil
			},
		}
		svc := NewAuthService(repo, jwtSecret)

		u := &userModel.User{Username: fakeUsername, Email: fakeValidEmail, Password: plainTextPassword}
		err := svc.Register(u)

		So(err, ShouldBeNil)
		So(CheckPasswordHash(plainTextPassword, u.Password), ShouldBeTrue)
	})

	Convey("Register should propagate create error", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{}, userModel.ErrUserNotFound
			},
			createFunc: func(user *userModel.User) error {
				return errors.New("insert failed")
			},
		}
		svc := NewAuthService(repo, jwtSecret)

		err := svc.Register(&userModel.User{Username: fakeUsername, Email: fakeValidEmail, Password: plainTextPassword})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrAuthRegisterFailed), ShouldBeTrue)
	})
}

func TestAuthServiceGenerateToken(t *testing.T) {
	Convey("GenerateToken should produce a valid signed JWT with expected claims", t, func() {
		svc := NewAuthService(&fakeUserRepo{}, jwtSecret)
		u := &userModel.User{ID: 9, Username: "alice"}

		tokenString, err := svc.GenerateToken(u)

		So(err, ShouldBeNil)
		So(tokenString, ShouldNotBeBlank)

		parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			So(token.Method.Alg(), ShouldEqual, jwt.SigningMethodHS256.Alg())
			return []byte(jwtSecret), nil
		})

		So(err, ShouldBeNil)
		So(parsed.Valid, ShouldBeTrue)

		claims := parsed.Claims.(jwt.MapClaims)
		So(int(claims["user_id"].(float64)), ShouldEqual, 9)
		So(claims["username"].(string), ShouldEqual, "alice")

		exp := int64(claims["exp"].(float64))
		So(exp, ShouldBeGreaterThan, time.Now().Unix())
	})
}

func TestAuthServiceLogin(t *testing.T) {
	Convey("Login should reject invalid email", t, func() {
		svc := NewAuthService(&fakeUserRepo{}, jwtSecret)

		accessToken, refreshToken, err := svc.Login(fakeBadEmail, plainTextPassword)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidEmailFormat), ShouldBeTrue)
		So(accessToken, ShouldEqual, "")
		So(refreshToken, ShouldEqual, "")
	})

	Convey("Login should propagate repo error", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{}, errors.New("db error")
			},
		}
		svc := NewAuthService(repo, jwtSecret)

		_, _, err := svc.Login(fakeValidEmail, plainTextPassword)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrAuthLoginFailed), ShouldBeTrue)
	})

	Convey("Login should fail on wrong password", t, func() {
		hash, hashErr := HashPassword(plainTextPassword)
		So(hashErr, ShouldBeNil)

		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{ID: 2, Username: "bob", Email: email, Password: hash}, nil
			},
		}
		svc := NewAuthService(repo, jwtSecret)

		_, _, err := svc.Login(fakeValidEmail, wrongPassword)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidCredentials), ShouldBeTrue)
	})

	Convey("Login should return invalid credentials when user is missing", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{}, userModel.ErrUserNotFound
			},
		}
		svc := NewAuthService(repo, jwtSecret)

		_, _, err := svc.Login(fakeValidEmail, plainTextPassword)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidCredentials), ShouldBeTrue)
	})

	Convey("Login should return access and refresh tokens for valid credentials", t, func() {
		hash, hashErr := HashPassword(plainTextPassword)
		So(hashErr, ShouldBeNil)

		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{ID: 7, Username: "dave", Email: email, Password: hash}, nil
			},
		}
		svc := NewAuthService(repo, jwtSecret)

		accessToken, refreshToken, err := svc.Login(fakeValidEmail, plainTextPassword)

		So(err, ShouldBeNil)
		So(accessToken, ShouldNotBeBlank)
		So(refreshToken, ShouldNotBeBlank)

		parsedAccess, parseAccessErr := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})
		So(parseAccessErr, ShouldBeNil)
		So(parsedAccess.Valid, ShouldBeTrue)

		accessClaims := parsedAccess.Claims.(jwt.MapClaims)
		So(accessClaims["type"].(string), ShouldEqual, "access")

		parsedRefresh, parseRefreshErr := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})
		So(parseRefreshErr, ShouldBeNil)
		So(parsedRefresh.Valid, ShouldBeTrue)

		refreshClaims := parsedRefresh.Claims.(jwt.MapClaims)
		So(refreshClaims["type"].(string), ShouldEqual, "refresh")
	})
}
