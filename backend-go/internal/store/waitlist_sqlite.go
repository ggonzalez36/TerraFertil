package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var ErrDuplicateEmail = errors.New("email already exists")

type WaitlistStore struct {
	db *sql.DB
}

func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set pragma journal_mode: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func NewWaitlistStore(db *sql.DB) (*WaitlistStore, error) {
	s := &WaitlistStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *WaitlistStore) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS waitlist_signups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    source TEXT NOT NULL DEFAULT 'landing',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate waitlist_signups: %w", err)
	}
	return nil
}

func (s *WaitlistStore) CreateSignup(email string, source string) (int64, time.Time, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(
		"INSERT INTO waitlist_signups(email, source, created_at) VALUES (?, ?, ?)",
		email,
		source,
		now,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, time.Time{}, ErrDuplicateEmail
		}
		return 0, time.Time{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, time.Time{}, err
	}

	return id, now, nil
}
