package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/VitoMinheere/TrackDate/internal/store"
)

type trackDayView struct {
	ID           int64
	Date         string
	Track        string
	Organization string
	Cost         string
}

func toTrackDayView(t store.TrackDay, currency string) trackDayView {
	return trackDayView{
		ID:           t.ID,
		Date:         t.Date,
		Track:        t.Track,
		Organization: t.Organization,
		Cost:         formatCost(t.CostCents, currency),
	}
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	days, err := s.store.ListTrackDaysByUser(user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	views := make([]trackDayView, 0, len(days))
	for _, d := range days {
		views = append(views, toTrackDayView(d, s.cfg.Currency))
	}

	share, err := s.store.ActiveShareLink(user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	render(w, http.StatusOK, "dashboard.html", map[string]any{
		"Email":     user.Email,
		"TrackDays": views,
		"ShareURL":  s.cfg.BaseURL + "/s/" + share.Slug,
	})
}

func (s *Server) handleCreateTrackDay(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	in, err := parseTrackDayForm(r)
	if err != nil {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}

	if _, err := s.store.CreateTrackDay(user.ID, in); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (s *Server) handleEditTrackDayForm(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	day, err := s.store.GetTrackDay(user.ID, id)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	render(w, http.StatusOK, "edit_trackday.html", map[string]any{
		"Email":     user.Email,
		"Day":       day,
		"CostInput": strconv.FormatFloat(float64(day.CostCents)/100, 'f', 2, 64),
	})
}

func (s *Server) handleUpdateTrackDay(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	in, err := parseTrackDayForm(r)
	if err != nil {
		http.Redirect(w, r, "/app/trackdays/"+r.PathValue("id")+"/edit", http.StatusSeeOther)
		return
	}

	if _, err := s.store.UpdateTrackDay(user.ID, id, in); errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (s *Server) handleDeleteTrackDay(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := s.store.DeleteTrackDay(user.ID, id); err != nil && !errors.Is(err, store.ErrNotFound) {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func parseTrackDayForm(r *http.Request) (store.TrackDayInput, error) {
	date := strings.TrimSpace(r.FormValue("date"))
	track := strings.TrimSpace(r.FormValue("track"))
	org := strings.TrimSpace(r.FormValue("organization"))

	if date == "" || track == "" || org == "" {
		return store.TrackDayInput{}, errors.New("missing required field")
	}

	cents, err := parseCost(r.FormValue("cost"))
	if err != nil {
		return store.TrackDayInput{}, err
	}

	return store.TrackDayInput{
		Date:         date,
		Track:        track,
		Organization: org,
		CostCents:    cents,
	}, nil
}
