package database

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

var (
	ErrFilepathExists = errors.New("filepath already exists")
)

type SyncFile struct {
	Id        uint64
	UserId    uint64
	Filepath  string
	Etag      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func UpdateSyncFileFilepath(db *sql.DB, currFilepath, newFilepath string) error {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM file_syncs WHERE filepath=?", newFilepath)
	if err := row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoResults
		} else {
			return err // unexpected error
		}
	}
	if count != 0 {
		return ErrFilepathExists
	}
	_, err := db.Exec("UPDATE file_syncs SET filepath=? WHERE filepath=?", newFilepath, currFilepath)
	if err != nil {
		return err
	}

	return nil
}

func UpdateSyncFileEtag(db *sql.DB, filepath, etag string) error {
	_, err := db.Exec("UPDATE file_syncs SET etag=? WHERE filepath=?", etag, filepath)
	if err != nil {
		return err
	}

	return nil
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
