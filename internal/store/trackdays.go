package store

import (
	"database/sql"
	"errors"
	"time"
)

type TrackDay struct {
	ID           int64
	UserID       int64
	Date         string // YYYY-MM-DD
	Track        string
	Organization string
	CostCents    int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TrackDayInput struct {
	Date         string
	Track        string
	Organization string
	CostCents    int64
}

func (s *Store) CreateTrackDay(userID int64, in TrackDayInput) (TrackDay, error) {
	res, err := s.db.Exec(
		`INSERT INTO track_days (user_id, date, track, organization, cost_cents) VALUES (?, ?, ?, ?, ?)`,
		userID, in.Date, in.Track, in.Organization, in.CostCents,
	)
	if err != nil {
		return TrackDay{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return TrackDay{}, err
	}
	return s.GetTrackDay(userID, id)
}

func (s *Store) GetTrackDay(userID, id int64) (TrackDay, error) {
	var t TrackDay
	err := s.db.QueryRow(
		`SELECT id, user_id, date, track, organization, cost_cents, created_at, updated_at
		 FROM track_days WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&t.ID, &t.UserID, &t.Date, &t.Track, &t.Organization, &t.CostCents, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return TrackDay{}, ErrNotFound
	}
	if err != nil {
		return TrackDay{}, err
	}
	return t, nil
}

// ListTrackDaysByUser returns a user's track days ordered soonest-first.
func (s *Store) ListTrackDaysByUser(userID int64) ([]TrackDay, error) {
	return s.queryTrackDays(`SELECT id, user_id, date, track, organization, cost_cents, created_at, updated_at
		FROM track_days WHERE user_id = ? ORDER BY date ASC, id ASC`, userID)
}

func (s *Store) queryTrackDays(query string, args ...any) ([]TrackDay, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []TrackDay
	for rows.Next() {
		var t TrackDay
		if err := rows.Scan(&t.ID, &t.UserID, &t.Date, &t.Track, &t.Organization, &t.CostCents, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		days = append(days, t)
	}
	return days, rows.Err()
}

func (s *Store) UpdateTrackDay(userID, id int64, in TrackDayInput) (TrackDay, error) {
	res, err := s.db.Exec(
		`UPDATE track_days SET date = ?, track = ?, organization = ?, cost_cents = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		in.Date, in.Track, in.Organization, in.CostCents, id, userID,
	)
	if err != nil {
		return TrackDay{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return TrackDay{}, err
	}
	if n == 0 {
		return TrackDay{}, ErrNotFound
	}
	return s.GetTrackDay(userID, id)
}

func (s *Store) DeleteTrackDay(userID, id int64) error {
	res, err := s.db.Exec(`DELETE FROM track_days WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
