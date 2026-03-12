package db

import (
	"database/sql"
	"fmt"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

var DB *sql.DB

var (
	getConfigString = func(key string) (string, error) {
		return beego.AppConfig.String(key)
	}
	openDB = func(driverName, dsn string) (*sql.DB, error) {
		return sql.Open(driverName, dsn)
	}
	pingDB = func(db *sql.DB) error {
		return db.Ping()
	}
)

// InitDB initializes the database connection using environment variables for configuration.
// It sets up the connection pool and tests the connection to ensure it's working properly.
// If any errors occur during the initialization process, it will panic with an appropriate error message.
func InitDB() {
	host, _ := getConfigString("database::DB_HOST")
	port, _ := getConfigString("database::DB_PORT")
	user, _ := getConfigString("database::DB_USER")
	password, _ := getConfigString("database::DB_PASSWORD")
	dbname, _ := getConfigString("database::DB_NAME")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)

	// Initialize the database connection using the DSN
	var err error
	DB, err = openDB("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the database: %v", err))
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the database connection
	err = pingDB(DB)

	if err != nil {
		panic(fmt.Sprintf("Failed to ping the database: %v", err))
	}
	fmt.Println("Database connection established successfully.")
}