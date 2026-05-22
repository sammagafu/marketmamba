package api

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"forex-bot/internal/auth"
	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
	"forex-bot/internal/users"
)

//go:embed dist/*
var staticFiles embed.FS

type BrokerResolver func(userID int64) (broker.Broker, error)

type Server struct {
	cfg           *config.Config
	storage       *storage.PostgresStorage
	subs          *subscription.Service
	users         *users.Service
	resolveBroker BrokerResolver
	mux           *http.ServeMux
}

func NewServer(cfg *config.Config, store *storage.PostgresStorage, subs *subscription.Service, usersSvc *users.Service, resolve BrokerResolver) *Server {
	s := &Server{
		cfg:           cfg,
		storage:       store,
		subs:          subs,
		users:         usersSvc,
		resolveBroker: resolve,
		mux:           http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/config", s.handlePublicConfig)
	s.mux.HandleFunc("/api/v1/auth/telegram", s.handleTelegramLogin)
	s.mux.HandleFunc("/api/v1/auth/telegram/oidc", s.handleTelegramOIDCLogin)
	s.mux.HandleFunc("/api/v1/auth/telegram/callback", s.handleTelegramLoginCallback)
	s.mux.HandleFunc("/api/v1/auth/email", s.handleEmailLogin)
	s.mux.HandleFunc("/api/v1/auth/me", s.withUser(s.handleAuthMe))
	s.mux.HandleFunc("/api/v1/brokers/types", s.withUser(s.handleBrokerTypes))
	s.mux.HandleFunc("/api/v1/brokers/connection", s.withUser(s.handleBrokerConnection))
	s.mux.HandleFunc("/api/v1/brokers/test", s.withUser(s.handleBrokerTest))
	s.mux.HandleFunc("/api/v1/status", s.withUser(s.handleStatus))
	s.mux.HandleFunc("/api/v1/account", s.withUser(s.handleAccount))
	s.mux.HandleFunc("/api/v1/positions", s.withUser(s.handlePositions))
	s.mux.HandleFunc("/api/v1/subscription", s.withUser(s.handleSubscription))
	s.mux.HandleFunc("/api/v1/admin/stats", s.withAdmin(s.handleAdminStats))
	s.mux.HandleFunc("/api/v1/admin/users", s.withAdmin(s.handleAdminUsers))
	s.mux.HandleFunc("/api/v1/admin/activate", s.withAdmin(s.handleAdminActivate))
	s.mux.HandleFunc("/api/v1/admin/users/block", s.withAdmin(s.handleAdminBlockUser))
	s.mux.HandleFunc("/api/v1/admin/users/revoke", s.withAdmin(s.handleAdminRevokeSubscription))

	s.registerStatic()
}

func (s *Server) registerStatic() {
	if !s.cfg.App.EnableWeb {
		return
	}
	staticFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		log.Printf("[web] embed dist failed: %v", err)
		return
	}
	if _, err := fs.Stat(staticFS, "index.html"); err != nil {
		log.Printf("[web] index.html missing in embed — run: make web-build")
		return
	}
	// GET /{$} only matches "/" in Go 1.22+ mux — assets need their own pattern.
	assets := http.FileServer(http.FS(staticFS))
	s.mux.HandleFunc("GET /assets/{path...}", func(w http.ResponseWriter, r *http.Request) {
		assets.ServeHTTP(w, r)
	})
	s.mux.HandleFunc("GET /{$}", spaHandler(staticFS))
	log.Printf("[web] dashboard static files ready")
}

func spaHandler(staticFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "/" {
			serveFile(w, staticFS, "index.html")
			return
		}
		if _, err := fs.Stat(staticFS, path); err == nil {
			serveFile(w, staticFS, path)
			return
		}
		serveFile(w, staticFS, "index.html")
	}
}

// serveSPA kept for tests / compatibility
func serveSPA(staticFS fs.FS) http.HandlerFunc {
	return spaHandler(staticFS)
}

func serveFile(w http.ResponseWriter, fsys fs.FS, name string) {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	if strings.HasSuffix(name, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if strings.HasSuffix(name, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(name, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}
	w.Write(b)
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && s.allowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Telegram-User-Id, Authorization")
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
	return func(w http.ResponseWriter, r *http.Request) {
		var uid int64
		var err error

		if token := bearerToken(r); token != "" {
			uid, err = auth.Verify(s.sessionSecret(), token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired session — log in again")
				return
			}
		} else {
			if s.cfg.App.WebAPIKey != "" {
				key := r.Header.Get("X-API-Key")
				if key == "" {
					key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
				}
				if key != s.cfg.App.WebAPIKey {
					writeError(w, http.StatusUnauthorized, "log in with Telegram or provide API key")
					return
				}
			}
			raw := r.Header.Get("X-Telegram-User-Id")
			if raw == "" {
				writeError(w, http.StatusUnauthorized, "log in with Telegram")
				return
			}
			uid, err = strconv.ParseInt(raw, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid telegram user id")
				return
			}
		}

		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next(w, r.WithContext(ctx))
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		t := strings.TrimPrefix(h, "Bearer ")
		if strings.Contains(t, ".") {
			return t
		}
	}
	return r.Header.Get("X-Session-Token")
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
