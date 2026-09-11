package store

import (
	"database/sql"
	"errors"
	"time"
)

type ShareLink struct {
	ID        int64
	UserID    int64
	Slug      string
	CreatedAt time.Time
}

// ActiveShareLink returns the user's current (non-revoked) share link,
// creating one on first use. Share links live in their own table (rather
// than a slug column on users) so that a user can hold more than one, each
// potentially scoped differently — needed once per-viewer visibility ships,
// without a schema rewrite.
func (s *Store) ActiveShareLink(userID int64) (ShareLink, error) {
	link, err := s.getActiveShareLink(userID)
	if err == nil {
		return link, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return ShareLink{}, err
	}
	return s.createShareLink(userID)
}

func (s *Store) getActiveShareLink(userID int64) (ShareLink, error) {
	var l ShareLink
	err := s.db.QueryRow(
		`SELECT id, user_id, slug, created_at FROM share_links WHERE user_id = ? AND revoked_at IS NULL ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&l.ID, &l.UserID, &l.Slug, &l.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ShareLink{}, ErrNotFound
	}
	if err != nil {
		return ShareLink{}, err
	}
	return l, nil
}

func (s *Store) createShareLink(userID int64) (ShareLink, error) {
	slug, err := newSlug()
	if err != nil {
		return ShareLink{}, err
	}
	res, err := s.db.Exec(`INSERT INTO share_links (user_id, slug) VALUES (?, ?)`, userID, slug)
	if err != nil {
		return ShareLink{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return ShareLink{}, err
	}
	return ShareLink{ID: id, UserID: userID, Slug: slug, CreatedAt: time.Now()}, nil
}

// RegenerateShareLink revokes the user's current share link and issues a
// new one, invalidating any previously shared URL.
func (s *Store) RegenerateShareLink(userID int64) (ShareLink, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return ShareLink{}, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE share_links SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`, time.Now(), userID); err != nil {
		return ShareLink{}, err
	}

	slug, err := newSlug()
	if err != nil {
		return ShareLink{}, err
	}
	res, err := tx.Exec(`INSERT INTO share_links (user_id, slug) VALUES (?, ?)`, userID, slug)
	if err != nil {
		return ShareLink{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return ShareLink{}, err
	}
	if err := tx.Commit(); err != nil {
		return ShareLink{}, err
	}
	return ShareLink{ID: id, UserID: userID, Slug: slug, CreatedAt: time.Now()}, nil
}

// UserForShareSlug resolves a public share slug to the owning user, for the
// read-only public view. Revoked links resolve to ErrNotFound.
func (s *Store) UserForShareSlug(slug string) (User, error) {
	var userID int64
	err := s.db.QueryRow(`SELECT user_id FROM share_links WHERE slug = ? AND revoked_at IS NULL`, slug).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return s.GetUserByID(userID)
}
