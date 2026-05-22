package api

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
)

//go:embed dist/*
var staticFiles embed.FS

type BrokerResolver func(userID int64) (broker.Broker, error)

type Server struct {
	cfg           *config.Config
	storage       *storage.PostgresStorage
	subs          *subscription.Service
	resolveBroker BrokerResolver
	mux           *http.ServeMux
}

func NewServer(cfg *config.Config, store *storage.PostgresStorage, subs *subscription.Service, resolve BrokerResolver) *Server {
	s := &Server{
		cfg:           cfg,
		storage:       store,
		subs:          subs,
		resolveBroker: resolve,
		mux:           http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/config", s.handlePublicConfig)
	s.mux.HandleFunc("/api/v1/brokers/types", s.withAuth(s.handleBrokerTypes))
	s.mux.HandleFunc("/api/v1/brokers/connection", s.withUser(s.handleBrokerConnection))
	s.mux.HandleFunc("/api/v1/brokers/test", s.withUser(s.handleBrokerTest))
	s.mux.HandleFunc("/api/v1/status", s.withUser(s.handleStatus))
	s.mux.HandleFunc("/api/v1/account", s.withUser(s.handleAccount))
	s.mux.HandleFunc("/api/v1/positions", s.withUser(s.handlePositions))
	s.mux.HandleFunc("/api/v1/subscription", s.withUser(s.handleSubscription))
	s.mux.HandleFunc("/api/v1/admin/stats", s.withAdmin(s.handleAdminStats))
	s.mux.HandleFunc("/api/v1/admin/users", s.withAdmin(s.handleAdminUsers))
	s.mux.HandleFunc("/api/v1/admin/activate", s.withAdmin(s.handleAdminActivate))

	if staticFS, err := fs.Sub(staticFiles, "dist"); err == nil {
		s.mux.Handle("/", spaHandler(staticFS))
	}
}

func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "/" {
			r.URL.Path = "/"
		} else if _, err := fs.Stat(staticFS, path); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && s.allowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Telegram-User-Id")
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

type ctxKey int

const userIDKey ctxKey = 1

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.App.WebAPIKey != "" {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			}
			if key != s.cfg.App.WebAPIKey {
				writeError(w, http.StatusUnauthorized, "invalid API key")
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) withUser(next http.HandlerFunc) http.HandlerFunc {
	return s.withAuth(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("X-Telegram-User-Id")
		if raw == "" {
			writeError(w, http.StatusBadRequest, "X-Telegram-User-Id header required")
			return
		}
		uid, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid telegram user id")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next(w, r.WithContext(ctx))
	})
}

func (s *Server) withAdmin(next http.HandlerFunc) http.HandlerFunc {
	return s.withUser(func(w http.ResponseWriter, r *http.Request) {
		adminID := r.Context().Value(userIDKey).(int64)
		if !s.cfg.IsAdmin(adminID) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	})
}

func userIDFrom(r *http.Request) int64 {
	return r.Context().Value(userIDKey).(int64)
}
