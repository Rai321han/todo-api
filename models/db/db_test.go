package db

import (
	"database/sql"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitDB(t *testing.T) {
	Convey("InitDB success path", t, func() {
		oldGet := getConfigString
		oldOpen := openDB
		oldPing := pingDB
		oldDB := DB
		defer func() {
			getConfigString = oldGet
			openDB = oldOpen
			pingDB = oldPing
			DB = oldDB
		}()

		config := map[string]string{
			"database::DB_HOST":     "localhost",
			"database::DB_PORT":     "5432",
			"database::DB_USER":     "postgres",
			"database::DB_PASSWORD": "secret",
			"database::DB_NAME":     "todo_db",
		}

		getConfigString = func(key string) (string, error) {
			return config[key], nil
		}

		fakeDB := &sql.DB{}
		openCalled := false
		pingCalled := false

		openDB = func(driverName, dsn string) (*sql.DB, error) {
			openCalled = true
			So(driverName, ShouldEqual, "postgres")
			So(dsn, ShouldEqual, "postgres://postgres:secret@localhost:5432/todo_db?sslmode=disable")
			return fakeDB, nil
		}

		pingDB = func(db *sql.DB) error {
			pingCalled = true
			So(db, ShouldEqual, fakeDB)
			return nil
		}

		So(func() { InitDB() }, ShouldNotPanic)
		So(openCalled, ShouldBeTrue)
		So(pingCalled, ShouldBeTrue)
		So(DB, ShouldEqual, fakeDB)
	})

	Convey("InitDB should panic when open fails", t, func() {
		oldGet := getConfigString
		oldOpen := openDB
		oldPing := pingDB
		oldDB := DB
		defer func() {
			getConfigString = oldGet
			openDB = oldOpen
			pingDB = oldPing
			DB = oldDB
		}()

		getConfigString = func(key string) (string, error) {
			return "x", nil
		}

		openDB = func(driverName, dsn string) (*sql.DB, error) {
			return nil, errors.New("open failed")
		}

		So(func() { InitDB() }, ShouldPanicWith, "Failed to connect to the database: open failed")
	})

	Convey("InitDB should panic when ping fails", t, func() {
		oldGet := getConfigString
		oldOpen := openDB
		oldPing := pingDB
		oldDB := DB
		defer func() {
			getConfigString = oldGet
			openDB = oldOpen
			pingDB = oldPing
			DB = oldDB
		}()

		getConfigString = func(key string) (string, error) {
			return "x", nil
		}

		fakeDB := &sql.DB{}
		openDB = func(driverName, dsn string) (*sql.DB, error) {
			return fakeDB, nil
		}

		pingDB = func(db *sql.DB) error {
			return errors.New("ping failed")
		}

		So(func() { InitDB() }, ShouldPanicWith, "Failed to ping the database: ping failed")
	})
}
