package user

import (
	"database/sql"
)

// UserRepository provides methods to interact with the users table in the database.
// It allows for retrieving user information and creating new users.
type UserRepository struct {
	DB *sql.DB
}

// GetUserByEmail retrieves a user from the database by their email address.
// It returns the user if found, or an error if the user is not found or if any other database error occurs.
func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.DB.QueryRow("SELECT id, username, email, password_hash FROM users WHERE email=$1", email).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return &User{}, ErrUserNotFound
		}
		return &User{}, err
	}
	return &user, nil
}

// Create inserts a new user into the database.
// It takes a pointer to a User struct as input and returns an error if the operation fails. If the user is created successfully, it returns nil.
func (r *UserRepository) Create(user *User) error {
	_, err := r.DB.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1,$2,$3)",
		user.Username, user.Email, user.Password)
	return err
}
