package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const SQL_PROVIDER = "sqlite3"
const DEFAULT_DB_PATH = "default-sqlite.db?mode=memory"

var db *sql.DB

// Get the module's database object. If the module's database object hasn't
// been set yet, create the module's database object and return it.
func GetDB() (*sql.DB, error) {
	var err error
	if db != nil {
		return db, err
	}

	db, err = NewDB(DEFAULT_DB_PATH)
	return db, err
}

// Set the module's database object.
func SetDB(newDB *sql.DB) {
	db = newDB
}

// Create a new database object, providing a SQLite connection string.
//
// For more information on SQLite connection strings, see
// https://github.com/mattn/go-sqlite3?tab=readme-ov-file#connection-string
func NewDB(connStr string) (*sql.DB, error) {
	path, err := getAbsoluteDbPath(connStr)
	if err != nil {
		return nil, err
	}
	return sql.Open(SQL_PROVIDER, path)
}

func getAbsoluteDbPath(path string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("file:%s", filepath.Join(cwd, path)), nil
}
