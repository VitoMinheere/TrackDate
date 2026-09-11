package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type User struct {
	ID        int64
	Email     string
	CreatedAt time.Time
}

// GetOrCreateUserByEmail normalizes the email and returns the existing user,
// or creates one. Login never distinguishes "unknown email" from "known
// email" to the caller, so this is safe to call on every login attempt.
func (s *Store) GetOrCreateUserByEmail(email string) (User, error) {
	email = normalizeEmail(email)

	u, err := s.getUserByEmail(email)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return User{}, err
	}

	res, err := s.db.Exec(`INSERT INTO users (email) VALUES (?)`, email)
	if err != nil {
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return s.GetUserByID(id)
}

func (s *Store) getUserByEmail(email string) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT id, email, created_at FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) GetUserByID(id int64) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT id, email, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Email, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
