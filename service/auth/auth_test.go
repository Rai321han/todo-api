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

func TestIsValidEmail(t *testing.T) {
	Convey("isValidEmail should validate format", t, func() {
		So(isValidEmail("user@example.com"), ShouldBeTrue)
		So(isValidEmail("bad-email"), ShouldBeFalse)
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
		svc := NewAuthService(&fakeUserRepo{}, "jwt-secret")

		err := svc.Register(&userModel.User{Email: "bad-email", Password: "pass"})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidEmailFormat), ShouldBeTrue)
	})

	Convey("Register should reject existing user", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				So(email, ShouldEqual, "existing@example.com")
				return &userModel.User{ID: 1, Email: email}, nil
			},
		}
		svc := NewAuthService(repo, "jwt-secret")

		err := svc.Register(&userModel.User{Email: "existing@example.com", Password: "pass"})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrUserAlreadyExists), ShouldBeTrue)
	})

	Convey("Register should hash password and create user", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{}, userModel.ErrUserNotFound
			},
			createFunc: func(user *userModel.User) error {
				So(user.Email, ShouldEqual, "new@example.com")
				So(user.Password, ShouldNotEqual, "plain-pass")
				So(CheckPasswordHash("plain-pass", user.Password), ShouldBeTrue)
				return nil
			},
		}
		svc := NewAuthService(repo, "jwt-secret")

		u := &userModel.User{Email: "new@example.com", Password: "plain-pass"}
		err := svc.Register(u)

		So(err, ShouldBeNil)
		So(CheckPasswordHash("plain-pass", u.Password), ShouldBeTrue)
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
		svc := NewAuthService(repo, "jwt-secret")

		err := svc.Register(&userModel.User{Email: "new@example.com", Password: "pass"})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrAuthRegisterFailed), ShouldBeTrue)
	})
}

func TestAuthServiceGenerateToken(t *testing.T) {
	Convey("GenerateToken should produce a valid signed JWT with expected claims", t, func() {
		svc := NewAuthService(&fakeUserRepo{}, "my-secret")
		u := &userModel.User{ID: 9, Username: "alice"}

		tokenString, err := svc.GenerateToken(u)

		So(err, ShouldBeNil)
		So(tokenString, ShouldNotBeBlank)

		parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			So(token.Method.Alg(), ShouldEqual, jwt.SigningMethodHS256.Alg())
			return []byte("my-secret"), nil
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
		svc := NewAuthService(&fakeUserRepo{}, "jwt-secret")

		token, err := svc.Login("bad-email", "pass")

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidEmailFormat), ShouldBeTrue)
		So(token, ShouldEqual, "")
	})

	Convey("Login should propagate repo error", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{}, errors.New("db error")
			},
		}
		svc := NewAuthService(repo, "jwt-secret")

		_, err := svc.Login("user@example.com", "pass")

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrAuthLoginFailed), ShouldBeTrue)
	})

	Convey("Login should fail on wrong password", t, func() {
		hash, hashErr := HashPassword("correct-pass")
		So(hashErr, ShouldBeNil)

		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{ID: 2, Username: "bob", Email: email, Password: hash}, nil
			},
		}
		svc := NewAuthService(repo, "jwt-secret")

		_, err := svc.Login("user@example.com", "wrong-pass")

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidCredentials), ShouldBeTrue)
	})

	Convey("Login should return invalid credentials when user is missing", t, func() {
		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{}, userModel.ErrUserNotFound
			},
		}
		svc := NewAuthService(repo, "jwt-secret")

		_, err := svc.Login("missing@example.com", "any-pass")

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidCredentials), ShouldBeTrue)
	})

	Convey("Login should return token for valid credentials", t, func() {
		hash, hashErr := HashPassword("correct-pass")
		So(hashErr, ShouldBeNil)

		repo := &fakeUserRepo{
			getUserByEmailFunc: func(email string) (*userModel.User, error) {
				return &userModel.User{ID: 7, Username: "dave", Email: email, Password: hash}, nil
			},
		}
		svc := NewAuthService(repo, "jwt-secret")

		token, err := svc.Login("user@example.com", "correct-pass")

		So(err, ShouldBeNil)
		So(token, ShouldNotBeBlank)

		parsed, parseErr := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("jwt-secret"), nil
		})
		So(parseErr, ShouldBeNil)
		So(parsed.Valid, ShouldBeTrue)
	})
}

