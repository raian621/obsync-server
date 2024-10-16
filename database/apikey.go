package database

import (
	"database/sql"
	"time"
)

type ApiKey struct {
	Id        uint64
	UserId    uint64
	Name      string
	Hash      string
	Active    bool
	CreatedAt time.Time
}

func CreateApiKey(db *sql.DB, userId uint64, name, key string) (*ApiKey, error) {
	hash, err := HashPassword(key)
	now := time.Now().UTC()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		"INSERT INTO api_keys (user_id, hash, active, created_at)"+
			"VALUES (:user_id, :hash, :active, :created_at)",
		sql.Named("user_id", userId),
		sql.Named("hash", hash),
		sql.Named("active", true),
		sql.Named("created_at", now),
	)
	if err != nil {
		return nil, err
	}
	var id uint64
	row := db.QueryRow("SELECT id FROM api_keys WHERE name=?", name)
	if err := row.Scan(&id); err != nil {
		return nil, err
	}

	return &ApiKey{
		Id:        id,
		UserId:    userId,
		Name:      name,
		Hash:      hash,
		Active:    true,
		CreatedAt: now,
	}, nil
}
