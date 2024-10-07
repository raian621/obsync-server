package database

import (
	"database/sql"
	"log"
	"strings"
)

type migration struct {
	name         string
	sqlStatement string
}

var migrations = []migration{
	{
		name: "CreateUsersTable",
		sqlStatement: strings.Join([]string{
			"CREATE TABLE users (",
			"  id       INTEGER      PRIMARY KEY AUTOINCREMENT,",
			"  username VARCHAR(100) UNIQUE NOT NULL,",
			"  email    VARCHAR(200) UNIQUE NOT NULL,",
			"  passhash VARCHAR(97)  NOT NULL",
			");"},
			"\n",
		),
  },
  {
    name: "CreateSessionsTable",
    sqlStatement: strings.Join([]string{
      "CREATE TABLE sessions (",
      "  id          INTEGER  PRIMARY KEY AUTOINCREMENT,",
      "  session_key CHAR(36) UNIQUE NOT NULL,",
      "  expires     TEXT     NOT NULL,",
      "  user_id     INTEGER  REFERENCES users(id) ON DELETE CASCADE",
      ");"},
      "\n",
    ),
  },
}

func CreateMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS migrations (\n" +
		"  id   INTEGER      PRIMARY KEY AUTOINCREMENT,\n" +
		"  name VARCHAR(100) UNIQUE NOT NULL\n" +
		");",
	)
	return err
}

func (m *migration) Apply(db *sql.DB) error {
	log.Printf("Applying migration `%s`...\n", m.name)
	log.Println(m.sqlStatement)
  
  // check if the migration needs to be applied, if not, return early
	row := db.QueryRow(
		"SELECT COUNT(*) FROM migrations WHERE name=?",
		m.name,
	)
  var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count == 1 {
		log.Println("Migration already applied...")
		return nil
	}
  
  // apply migration
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(m.sqlStatement); err != nil {
		return err
	}
	if _, err := tx.Exec(
		"INSERT INTO migrations (name) VALUES (?)",
		m.name,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func ApplyMigrations(db *sql.DB) error {
	if err := CreateMigrationsTable(db); err != nil {
		return err
	}
	for _, migration := range migrations {
		if err := migration.Apply(db); err != nil {
			return err
		}
	}
	return nil
}
