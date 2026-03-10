package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"time"
	"todo-api/models/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService provides methods for user registration and authentication.
// It interacts with the UserRepository to manage user data and uses JWT for token generation.
type AuthService struct {
	repo UserRepo
	JwtSecret string
}

// UserRepo defines repository behavior required by AuthService.
type UserRepo interface {
	GetUserByEmail(email string) (*user.User, error)
	Create(user *user.User) error
}

// isValidEmail checks if the provided email address is in a valid format.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// NewAuthService creates a new instance of AuthService with the provided UserRepository and JWT secret key.
func NewAuthService(repo UserRepo, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, JwtSecret: jwtSecret}
}

// Register registers a new user by validating the email format, 
// checking for existing users, hashing the password, and creating the user in the database.
// It returns an error if any of these steps fail.
func (s *AuthService) Register(user *user.User) error {
	if !isValidEmail(user.Email) {
		return fmt.Errorf("invalid email format")
	}
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

// GenerateToken generates a JWT token for the authenticated user.
// It includes the user's ID and username in the token claims and sets an expiration time of 24 hours.
// The token is signed using the secret key provided in the AuthService.
func (s *AuthService) GenerateToken(user *user.User) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "username": user.Username,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.JwtSecret))
}


// Login authenticates a user by validating the email format, retrieving the user from the database,
// checking the password hash, and generating a JWT token if the credentials are valid.
// It returns the generated token or an error if any of the steps fail.
func (s *AuthService) Login(email, password string) (string, error) {
	// check if email is valid or not
	if !isValidEmail(email) {
		return "",fmt.Errorf("invalid email format")
	}

    user, err := s.repo.GetUserByEmail(email)
    
	if err != nil {
        return "", err
    }

    if !CheckPasswordHash(password, user.Password) {
        return "", errors.New("invalid credentials")
    }

	token, err := s.GenerateToken(user)
	if err != nil {
		return "", err
	}
    return token, nil
}