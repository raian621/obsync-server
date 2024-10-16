package database

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrUsernameFormat = errors.New("username too short or too long")
	ErrEmailFormat    = errors.New("email format invalid")
	ErrPasswordLength = errors.New("password is too short (must be at least 8 characters)")
)

type User struct {
	Id       uint64
	Username string
	Passhash string
	Email    string
}

func CreateUser(db *sql.DB, username string, email string, password string) (*User, error) {
	if len(username) == 0 || len(username) > 100 {
		return nil, ErrUsernameFormat
	}
	if !validEmail(email) || len(email) > 200 {
		return nil, ErrEmailFormat
	}
	if len(password) < 8 {
		return nil, ErrPasswordLength
	}

	passhash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := User{
		Username: username,
		Email:    email,
		Passhash: passhash,
	}

	_, err = db.Exec(
		strings.Join([]string{
			"INSERT INTO users (username, email, passhash)",
			"  VALUES (:username, :email, :passhash)"},
			"\n",
		),
		sql.Named("username", username),
		sql.Named("email", email),
		sql.Named("passhash", passhash),
	)
	if err != nil {
		return nil, err
	}
	row := db.QueryRow(
		"SELECT id FROM users WHERE username=:username",
		sql.Named("username", username),
	)
	if err := row.Scan(&user.Id); err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	if len(username) == 0 || len(username) > 100 {
		return nil, ErrUsernameFormat
	}

	row := db.QueryRow("SELECT id, username, email, passhash FROM users WHERE username=?", username)
	user := User{}
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.Passhash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	if !validEmail(email) || len(email) > 200 {
		return nil, ErrEmailFormat
	}

	row := db.QueryRow("SELECT id, username, email, passhash FROM users WHERE email=?", email)
	user := User{}
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.Passhash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserById(db *sql.DB, id uint64) (*User, error) {
	row := db.QueryRow("SELECT id, username, email, passhash FROM users WHERE id=?", id)
	user := User{}
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.Passhash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
