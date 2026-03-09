package auth

import (
	"errors"
	"fmt"
	"time"
	"todo-api/models/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *user.UserRepository
	JwtSecret string
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func NewAuthService(repo *user.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, JwtSecret: jwtSecret}
}

func (s *AuthService) Register(user *user.User) error {
	// check if user already exists or not
	_, err := s.repo.GetUserByEmail(user.Email)
	
	if err == nil {
		return fmt.Errorf("user already exists")
	}

	hashedPassword, err := HashPassword(user.Password)
	
	if err != nil {
		return err
	}
	
	user.Password = hashedPassword
	return s.repo.Create(user)
}

func (s *AuthService) GenerateToken(user *user.User) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "username": user.Username,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.JwtSecret))
}

func (s *AuthService) Login(email, password string) (string, error) {
    user, err := s.repo.GetUserByEmail(email)
    if err != nil {
        return "", err
    }
    if !CheckPasswordHash(password, user.Password) {
        return "", errors.New("invalid credentials")
    }
    return s.GenerateToken(user)
}