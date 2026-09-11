package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"github.com/VitoMinheere/TrackDate/internal/store"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if _, ok := userFromContext(r.Context()); ok {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	render(w, http.StatusOK, "login.html", nil)
}

func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" || !strings.Contains(email, "@") {
		render(w, http.StatusUnprocessableEntity, "login.html", map[string]any{
			"Error": "Enter a valid email address.",
		})
		return
	}

	user, err := s.store.GetOrCreateUserByEmail(email)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rawToken, ok, err := s.store.CreateMagicLink(user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if ok {
		verifyURL := s.cfg.BaseURL + "/login/verify?token=" + rawToken
		if err := s.mailer.SendMagicLink(user.Email, verifyURL); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
	// Same response whether or not a link was actually sent (cooldown hit,
	// or email unknown) — avoids leaking account existence or rate state.

	render(w, http.StatusOK, "check_email.html", nil)
}

func (s *Server) handleLoginVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		render(w, http.StatusBadRequest, "login_error.html", nil)
		return
	}

	userID, err := s.store.ConsumeMagicLink(token)
	if errors.Is(err, store.ErrNotFound) {
		render(w, http.StatusBadRequest, "login_error.html", nil)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sessionToken, err := s.store.CreateSession(userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	s.setSessionCookie(w, sessionToken)
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		_ = s.store.DeleteSession(cookie.Value)
	}
	s.clearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
