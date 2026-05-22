package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/storage"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	cfg       *config.Config
	storage   *storage.PostgresStorage
	broker    broker.Broker
	userID    int64
	mux       *http.ServeMux
}

func NewServer(cfg *config.Config, store *storage.PostgresStorage, b broker.Broker, primaryUserID int64) *Server {
	s := &Server{
		cfg:     cfg,
		storage: store,
		broker:  b,
		userID:  primaryUserID,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/brokers/types", s.withAuth(s.handleBrokerTypes))
	s.mux.HandleFunc("/api/v1/brokers/connection", s.withAuth(s.handleBrokerConnection))
	s.mux.HandleFunc("/api/v1/brokers/test", s.withAuth(s.handleBrokerTest))
	s.mux.HandleFunc("/api/v1/status", s.withAuth(s.handleStatus))
	s.mux.HandleFunc("/api/v1/account", s.withAuth(s.handleAccount))
	s.mux.HandleFunc("/api/v1/positions", s.withAuth(s.handlePositions))

	if staticFS, err := fs.Sub(staticFiles, "static"); err == nil {
		s.mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && s.allowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		s.mux.ServeHTTP(w, r)
	})
}

func (s *Server) allowedOrigin(origin string) bool {
	for _, o := range s.cfg.App.CORSOrigins {
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}
