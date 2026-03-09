package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection using environment variables for configuration. It sets up the connection pool and tests the connection to ensure it's working properly. If any errors occur during the initialization process, it will panic with an appropriate error message.
func InitDB() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")



	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)

	fmt.Println(dsn)

	// Initialize the database connection using the DSN
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the database: %v", err))
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the database connection
	err = DB.Ping()

	if err != nil {
		panic(fmt.Sprintf("Failed to ping the database: %v", err))
	}
	fmt.Println("Database connection established successfully.")
}