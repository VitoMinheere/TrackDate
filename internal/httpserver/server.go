// Package httpserver implements Trackdate's HTTP handlers, session
// middleware, and server-rendered HTML views.
package httpserver

import (
	"log"
	"net/http"
	"time"

	"github.com/VitoMinheere/TrackDate/internal/config"
	"github.com/VitoMinheere/TrackDate/internal/mail"
	"github.com/VitoMinheere/TrackDate/internal/store"
)

const sessionCookieName = "trackdate_session"

type Server struct {
	cfg    config.Config
	store  *store.Store
	mailer mail.Mailer
	mux    *http.ServeMux
}

func New(cfg config.Config, s *store.Store, mailer mail.Mailer) *Server {
	srv := &Server{cfg: cfg, store: s, mailer: mailer, mux: http.NewServeMux()}
	srv.routes()
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.withUser(s.mux).ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.Handle("GET /static/", http.FileServerFS(staticFS))

	s.mux.HandleFunc("GET /{$}", s.handleIndex)
	s.mux.HandleFunc("POST /login", s.handleLoginSubmit)
	s.mux.HandleFunc("GET /login/verify", s.handleLoginVerify)
	s.mux.HandleFunc("POST /logout", s.handleLogout)

	s.mux.HandleFunc("GET /app", s.requireAuth(s.handleDashboard))
	s.mux.HandleFunc("POST /app/trackdays", s.requireAuth(s.handleCreateTrackDay))
	s.mux.HandleFunc("GET /app/trackdays/{id}/edit", s.requireAuth(s.handleEditTrackDayForm))
	s.mux.HandleFunc("POST /app/trackdays/{id}/update", s.requireAuth(s.handleUpdateTrackDay))
	s.mux.HandleFunc("POST /app/trackdays/{id}/delete", s.requireAuth(s.handleDeleteTrackDay))
	s.mux.HandleFunc("POST /app/share/regenerate", s.requireAuth(s.handleRegenerateShare))

	s.mux.HandleFunc("GET /s/{slug}", s.handleShareView)
}

// StartJanitor periodically prunes expired login/session tokens so the
// tables don't grow unbounded on a long-running instance.
func (s *Server) StartJanitor(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.store.PruneExpiredMagicLinks(); err != nil {
					log.Printf("janitor: prune magic links: %v", err)
				}
				if err := s.store.PruneExpiredSessions(); err != nil {
					log.Printf("janitor: prune sessions: %v", err)
				}
			case <-stop:
				return
			}
		}
	}()
}
