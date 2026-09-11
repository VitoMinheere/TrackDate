package httpserver

import (
	"errors"
	"net/http"

	"github.com/VitoMinheere/TrackDate/internal/store"
)

func (s *Server) handleRegenerateShare(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	if _, err := s.store.RegenerateShareLink(user.ID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

// handleShareView renders a user's track days read-only, resolved by an
// unguessable share-link slug. No login required.
func (s *Server) handleShareView(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	owner, err := s.store.UserForShareSlug(slug)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	days, err := s.store.ListTrackDaysByUser(owner.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	views := make([]trackDayView, 0, len(days))
	for _, d := range days {
		views = append(views, toTrackDayView(d, s.cfg.Currency))
	}

	render(w, http.StatusOK, "share.html", map[string]any{
		"TrackDays": views,
	})
}
