package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
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

// Custom auth errors provide predictable handling across service and controller layers.
var (
	ErrInvalidAuthInput    = errors.New("invalid auth input")
	ErrInvalidEmailFormat  = errors.New("invalid email format")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAuthRegisterFailed  = errors.New("failed to register user")
	ErrAuthLoginFailed     = errors.New("failed to authenticate user")
	ErrTokenGenerationFail = errors.New("failed to generate auth token")
)

const tokenTTL = 24 * time.Hour

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

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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
func (s *AuthService) Register(newUser *user.User) error {
	if newUser == nil {
		return fmt.Errorf("%w: user payload is required", ErrInvalidAuthInput)
	}

	newUser.Username = strings.TrimSpace(newUser.Username)
	newUser.Email = normalizeEmail(newUser.Email)
	if newUser.Username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidAuthInput)
	}
	if newUser.Password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidAuthInput)
	}
	if !isValidEmail(newUser.Email) {
		return ErrInvalidEmailFormat
	}

	_, err := s.repo.GetUserByEmail(newUser.Email)
	if err == nil {
		return ErrUserAlreadyExists
	}
	if !errors.Is(err, user.ErrUserNotFound) {
		return fmt.Errorf("%w: %v", ErrAuthRegisterFailed, err)
	}

	hashedPassword, err := HashPassword(newUser.Password)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAuthRegisterFailed, err)
	}

	newUser.Password = hashedPassword
	if err := s.repo.Create(newUser); err != nil {
		return fmt.Errorf("%w: %v", ErrAuthRegisterFailed, err)
	}

	return nil
}

// GenerateToken generates a JWT token for the authenticated user.
// It includes the user's ID and username in the token claims and sets an expiration time of 24 hours.
// The token is signed using the secret key provided in the AuthService.
func (s *AuthService) GenerateToken(user *user.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("%w: user payload is required", ErrInvalidAuthInput)
	}
	if strings.TrimSpace(s.JwtSecret) == "" {
		return "", fmt.Errorf("%w: jwt secret is not configured", ErrTokenGenerationFail)
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(tokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.JwtSecret))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGenerationFail, err)
	}

	return tokenString, nil
}


// Login authenticates a user by validating the email format, retrieving the user from the database,
// checking the password hash, and generating a JWT token if the credentials are valid.
// It returns the generated token or an error if any of the steps fail.
func (s *AuthService) Login(email, password string) (string, error) {
	normalizedEmail := normalizeEmail(email)
	if !isValidEmail(normalizedEmail) {
		return "", ErrInvalidEmailFormat
	}
	if strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("%w: password is required", ErrInvalidAuthInput)
	}

	foundUser, err := s.repo.GetUserByEmail(normalizedEmail)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf("%w: %v", ErrAuthLoginFailed, err)
	}

	if !CheckPasswordHash(password, foundUser.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := s.GenerateToken(foundUser)
	if err != nil {
		return "", err
	}

	return token, nil
}