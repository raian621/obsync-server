package database

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

const (
	ApiKeyLength uint = 16
)

var (
	ErrApiKeyExists = errors.New("api key with specified name already exists")
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

	// check if an API key with the same name exists, return an error if so
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE name=?", name)
	if err := row.Scan(&count); err != nil {
		return nil, err
	}
	if count != 0 {
		return nil, ErrApiKeyExists
	}

	// create api key in the database
	_, err = db.Exec(
		"INSERT INTO api_keys (user_id, name, hash, active, created_at)"+
			"VALUES (:user_id, :name, :hash, :active, :created_at)",
		sql.Named("user_id", userId),
		sql.Named("name", name),
		sql.Named("hash", hash),
		sql.Named("active", true),
		sql.Named("created_at", now),
	)
	if err != nil {
		return nil, err
	}
	var id uint64
	row = db.QueryRow("SELECT id FROM api_keys WHERE name=?", name)
	if err := row.Scan(&id); err != nil {
		log.Println(err)
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

func DeleteApiKey(db *sql.DB, userId uint64, name string) error {
	_, err := db.Exec("DELETE FROM api_keys WHERE user_id=? AND name=?", userId, name)

	return err
}

func SetApiKeyActivation(db *sql.DB, userId uint64, name string, active bool) error {
	_, err := db.Exec(
		"UPDATE api_keys SET active=? WHERE user_id=? AND name=?",
		active,
		userId,
		name,
	)

	return err
}

func GetApiKeys(db *sql.DB, userId uint64) ([]*ApiKey, error) {
	rows, err := db.Query(
		"SELECT id, user_id, name, hash, active, created_at"+
			"FROM api_keys WHERE user_id=?",
		userId,
	)
	if err != nil {
		return nil, err
	}
	var apiKeys []*ApiKey
	for rows.Next() {
		var apiKey ApiKey
		err := rows.Scan(
			&apiKey.Id,
			&apiKey.UserId,
			&apiKey.Name,
			&apiKey.Hash,
			&apiKey.Active,
			&apiKey.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, &apiKey)
	}

	return apiKeys, nil
}
