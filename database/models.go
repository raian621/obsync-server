package database

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"
)

var (
	ErrUsernameFormat       = errors.New("username too short or too long")
	ErrEmailFormat          = errors.New("email format invalid")
	ErrPasswordLength       = errors.New("password is too short (must be at least 8 characters)")
	ErrNoResults            = errors.New("query returned no results")
	ErrIncorrectCredentials = errors.New("username or password incorrect")

	sessionTime     = 72 * time.Hour
	sessionKeyBytes = 16
)

type User struct {
	Id       uint64
	Username string
	Passhash string
	Email    string
}

type Session struct {
	Id         uint64
	SessionKey string
	UserId     uint64
	Expires    time.Time
}

type SyncFile struct {
	Id        uint64
	UserId    uint64
	Filepath  string
	Etag      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ApiKey struct {
	Id        uint64
	UserId    uint64
	Name      string
	Hash      string
	Active    bool
	CreatedAt time.Time
}

/// User model functions

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

/// Session model functions

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
	if err := ValidatePassword(password, passhash); err != nil {
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
		sessionKey, err = generateSessionKey(uint(sessionKeyBytes))
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

/// SyncFile model functions

func CreateSyncFile(
	db *sql.DB,
	filepath, etag string,
	userId uint64,
) (*SyncFile, error) {
	var syncFile SyncFile

	createdAt := time.Now().UTC()
	_, err := db.Exec(
		"INSERT INTO file_syncs (filepath, etag, created_at, updated_at, user_id)\n"+
			"  VALUES (:filepath, :etag, :created_at, :updated_at, :user_id)",
		sql.Named("filepath", filepath),
		sql.Named("etag", etag),
		sql.Named("created_at", createdAt),
		sql.Named("updated_at", createdAt),
		sql.Named("user_id", userId),
	)
	if err != nil {
		return nil, err
	}
	row := db.QueryRow("SELECT id FROM file_syncs WHERE filepath=?", filepath)
	if err := row.Scan(&syncFile.Id); err != nil {
		return nil, err
	}
	syncFile.Filepath = filepath
	syncFile.UserId = userId
	syncFile.Etag = etag
	syncFile.CreatedAt = createdAt
	syncFile.UpdatedAt = createdAt

	return &syncFile, nil
}

func GetSyncFileById(db *sql.DB, id uint64) (*SyncFile, error) {
	row := db.QueryRow(
		"SELECT id, user_id, filepath, etag, created_at, updated_at "+
			"FROM file_syncs WHERE id=?",
		id,
	)

	return scanSyncFile(row)
}

func GetSyncFileByFilepath(db *sql.DB, filepath string) (*SyncFile, error) {
	row := db.QueryRow(
		"SELECT id, user_id, filepath, etag, created_at, updated_at "+
			"FROM file_syncs WHERE filepath=?",
		filepath,
	)

	return scanSyncFile(row)
}

func GetSyncFilesByUserId(db *sql.DB, userId uint64) ([]*SyncFile, error) {
	var syncfiles []*SyncFile

	rows, err := db.Query(
		"SELECT id, user_id, filepath, etag, created_at, updated_at "+
			"FROM file_syncs WHERE user_id=?",
		userId,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		syncfile, err := scanSyncFile(rows)
		log.Println(syncfile)
		if err != nil {
			return nil, err
		}
		syncfiles = append(syncfiles, syncfile)
	}
	if len(syncfiles) == 0 {
		return nil, ErrNoResults
	}

	return syncfiles, nil
}

func scanSyncFile(row Scannable) (*SyncFile, error) {
	var (
		syncfile  SyncFile
		createdAt string
		updatedAt string
	)

	err := row.Scan(
		&syncfile.Id,
		&syncfile.UserId,
		&syncfile.Filepath,
		&syncfile.Etag,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoResults
		}
		return nil, err
	}
	syncfile.CreatedAt, err = time.Parse(ISO_8601_FORMAT, createdAt)
	if err != nil {
		return nil, err
	}
	syncfile.UpdatedAt, err = time.Parse(ISO_8601_FORMAT, updatedAt)
	if err != nil {
		return nil, err
	}

	return &syncfile, nil
}

// ApiKey model functions
