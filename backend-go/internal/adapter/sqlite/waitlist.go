package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"terrafertil/backend-go/internal/domain"
)

type WaitlistRepository struct {
	db *sql.DB
}

func NewWaitlistRepository(db *sql.DB) (*WaitlistRepository, error) {
	repo := &WaitlistRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *WaitlistRepository) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS waitlist_signups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    source TEXT NOT NULL DEFAULT 'landing',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_waitlist_email ON waitlist_signups(email);
`
	_, err := r.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate waitlist_signups: %w", err)
	}
	return nil
}

func (r *WaitlistRepository) Save(ctx context.Context, email, source string) (int64, time.Time, error) {
	now := time.Now().UTC()
	query := "INSERT INTO waitlist_signups(email, source, created_at) VALUES (?, ?, ?)"
	res, err := r.db.ExecContext(ctx, query, email, source, now)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, time.Time{}, domain.ErrDuplicateEmail
		}
		return 0, time.Time{}, fmt.Errorf("insert waitlist signup: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("retrieve last insert id: %w", err)
	}

	return id, now, nil
}

func (r *WaitlistRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM waitlist_signups WHERE email = ?)"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email existence: %w", err)
	}
	return exists, nil
}
