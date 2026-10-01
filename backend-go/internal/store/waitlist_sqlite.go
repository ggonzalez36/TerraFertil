package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"terrafertil/backend-go/internal/adapter/sqlite"
	"terrafertil/backend-go/internal/domain"
)

var ErrDuplicateEmail = domain.ErrDuplicateEmail

type WaitlistStore struct {
	repo *sqlite.WaitlistRepository
}

func OpenSQLite(path string) (*sql.DB, error) {
	// Enable WAL and 5s busy timeout to handle concurrent transactions without locking errors
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// SQLite single-writer pool configuration
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func NewWaitlistStore(db *sql.DB) (*sqlite.WaitlistRepository, error) {
	return sqlite.NewWaitlistRepository(db)
}
