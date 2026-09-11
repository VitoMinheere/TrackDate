// Command trackdate runs the Trackdate server: a single self-hosted binary
// backed by a SQLite file, for planning and sharing track days.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/VitoMinheere/TrackDate/internal/config"
	"github.com/VitoMinheere/TrackDate/internal/httpserver"
	"github.com/VitoMinheere/TrackDate/internal/mail"
	"github.com/VitoMinheere/TrackDate/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	s, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer s.Close()

	var mailer mail.Mailer
	if cfg.SMTPHost == "" {
		log.Println("SMTP_HOST not set; magic links will be logged to stdout instead of emailed")
		mailer = mail.ConsoleMailer{}
	} else {
		mailer = &mail.SMTPMailer{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
		}
	}

	srv := httpserver.New(cfg, s, mailer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	janitorStop := make(chan struct{})
	srv.StartJanitor(1*time.Hour, janitorStop)
	defer close(janitorStop)

	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("trackdate listening on %s (base url %s)", cfg.Addr, cfg.BaseURL)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
