package store

import (
	"database/sql"
	"errors"
	"time"
)

const sessionTTL = 30 * 24 * time.Hour

func (s *Store) CreateSession(userID int64) (rawToken string, err error) {
	raw, hash, err := newToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(
		`INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, hash, time.Now().Add(sessionTTL),
	)
	if err != nil {
		return "", err
	}
	return raw, nil
}

// UserForSession resolves a raw session token to its user, if the session
// exists and hasn't expired.
func (s *Store) UserForSession(rawToken string) (User, error) {
	hash := hashToken(rawToken)

	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE token_hash = ?`, hash).
		Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	if time.Now().After(expiresAt) {
		return User{}, ErrNotFound
	}
	return s.GetUserByID(userID)
}

func (s *Store) DeleteSession(rawToken string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hashToken(rawToken))
	return err
}

func (s *Store) PruneExpiredSessions() error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	return err
}
