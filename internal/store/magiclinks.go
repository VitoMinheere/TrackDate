package store

import (
	"database/sql"
	"errors"
	"time"
)

const magicLinkTTL = 15 * time.Minute

// magicLinkCooldown is the minimum gap between two magic-link requests for
// the same user, to make the login form a poor spam vector.
const magicLinkCooldown = 60 * time.Second

// CreateMagicLink issues a new login token for the user, unless one was
// already issued within the cooldown window (in which case ok is false and
// no email should be sent).
func (s *Store) CreateMagicLink(userID int64) (rawToken string, ok bool, err error) {
	var lastCreated time.Time
	err = s.db.QueryRow(`SELECT created_at FROM magic_links WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).
		Scan(&lastCreated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	if err == nil && time.Since(lastCreated) < magicLinkCooldown {
		return "", false, nil
	}

	raw, hash, err := newToken()
	if err != nil {
		return "", false, err
	}
	_, err = s.db.Exec(
		`INSERT INTO magic_links (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, hash, time.Now().Add(magicLinkTTL),
	)
	if err != nil {
		return "", false, err
	}
	return raw, true, nil
}

// ConsumeMagicLink validates and single-use-consumes a raw token, returning
// the associated user ID. It fails for unknown, expired, or already-used
// tokens.
func (s *Store) ConsumeMagicLink(rawToken string) (userID int64, err error) {
	hash := hashToken(rawToken)

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var (
		id         int64
		uid        int64
		expiresAt  time.Time
		consumedAt sql.NullTime
	)
	err = tx.QueryRow(
		`SELECT id, user_id, expires_at, consumed_at FROM magic_links WHERE token_hash = ?`, hash,
	).Scan(&id, &uid, &expiresAt, &consumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if consumedAt.Valid {
		return 0, ErrNotFound
	}
	if time.Now().After(expiresAt) {
		return 0, ErrNotFound
	}

	if _, err := tx.Exec(`UPDATE magic_links SET consumed_at = ? WHERE id = ?`, time.Now(), id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uid, nil
}

// PruneExpiredMagicLinks deletes stale rows so the table doesn't grow
// unbounded on a long-running instance.
func (s *Store) PruneExpiredMagicLinks() error {
	_, err := s.db.Exec(`DELETE FROM magic_links WHERE expires_at < ? OR consumed_at IS NOT NULL`, time.Now().Add(-24*time.Hour))
	return err
}
