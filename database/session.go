package database

import (
	"database/sql"
	"time"
)

var (
	sessionTime     = 72 * time.Hour
	sessionKeyBytes = 16
)

type Session struct {
	Id         uint64
	SessionKey string
	UserId     uint64
	Expires    time.Time
}

func LoginUser(db *sql.DB, username, password string) (*Session, error) {
	var (
		userId   uint64
		passhash string
	)

	row := db.QueryRow("SELECT id, passhash FROM users WHERE username=?", username)
	if err := row.Scan(&userId, &passhash); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrIncorrectCredentials
		}
		return nil, err // unexpected error
	}
	if err := ValidateHash(password, passhash); err != nil {
		if err == ErrInvalidPassword {
			return nil, ErrIncorrectCredentials
		}
		return nil, err
	}

	return CreateSession(db, userId)
}

func LogoutUser(db *sql.DB, sessionKey string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE session_key=?", sessionKey)
	return err
}

func CreateSession(db *sql.DB, userId uint64) (*Session, error) {
	var (
		sessionKey string
		err        error
		count      = 1
	)

	// make sure session keys aren't duplicated. it's very unlikely there will
	// ever be duplicate session keys, but I'd rather be correct than unlucky.
	for count > 0 {
		sessionKey, err = generateKey(uint(sessionKeyBytes))
		if err != nil {
			return nil, err
		}
		row := db.QueryRow(
			"SELECT COUNT(*) FROM sessions WHERE session_key=?",
			sessionKey,
		)
		if err := row.Scan(&count); err != nil {
			return nil, err
		}
	}

	expires := time.Now().Add(sessionTime)
	_, err = db.Exec(
		"INSERT INTO sessions (user_id, session_key, expires)\n"+
			"  VALUES (:user_id, :session_key, :expires)",
		sql.Named("user_id", userId),
		sql.Named("session_key", sessionKey),
		sql.Named("expires", expires),
	)
	if err != nil {
		return nil, err
	}

	var id uint64
	row := db.QueryRow("SELECT id FROM sessions WHERE session_key=?", sessionKey)
	if err := row.Scan(&id); err != nil {
		return nil, err
	}

	return &Session{
		Id:         id,
		UserId:     userId,
		SessionKey: sessionKey,
		Expires:    expires,
	}, nil
}

func DeleteExpiredSessions(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sessions WHERE expires<datetime('now')")
	return err
}