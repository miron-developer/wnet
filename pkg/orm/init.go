package orm

import (
	"context"
	"database/sql"
	"log"

	// sqlite driver
	_ "github.com/mattn/go-sqlite3"

	// db migration
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
)

// ConnToDB is var to conn to db
var ConnToDB *sql.DB

func warn(eLogger *log.Logger, e error) {
	if e != nil {
		eLogger.Fatal(e)
	}
}

func execMigrations(eLogger *log.Logger) error {
	driver, e := sqlite3.WithInstance(ConnToDB, &sqlite3.Config{})
	if e != nil {
		return e
	}

	m, e := migrate.NewWithDatabaseInstance("file://db/migrations", "sqlite3", driver)
	if e != nil {
		return e
	}

	if e = m.Up(); e != nil && e != migrate.ErrNoChange {
		return e
	}

	return nil
}

// InitDB init db, settings and tables
func InitDB(eLogger *log.Logger) {
	initGoogleDriveFileService(eLogger)
	getDBFileFromGoogleDrive(eLogger)

	ConnToDB, _ = sql.Open("sqlite3", "file:db/wnet.db?_auth&_auth_user=wnet&_auth_pass=wnet&_auth_crypt=sha1")

	_, e := ConnToDB.ExecContext(context.Background(), "PRAGMA foreign_keys = ON;PRAGMA case_sensitive_like = true;PRAGMA auto_vacuum = FULL;")
	warn(eLogger, e)

	warn(eLogger, execMigrations(eLogger))
}
