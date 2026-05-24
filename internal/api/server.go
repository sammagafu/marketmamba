package api

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/pairs"
	"forex-bot/internal/payments"
	"forex-bot/internal/risk"
	"forex-bot/internal/signals"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
	"forex-bot/internal/tier"
	"forex-bot/internal/users"
)

//go:embed dist/*
var staticFiles embed.FS

type BrokerResolver func(userID int64) (broker.Broker, error)

type Server struct {
	cfg           *config.Config
	storage       *storage.PostgresStorage
	subs          *subscription.Service
	tier          *tier.Service
	payments      *payments.Service
	users         *users.Service
	resolveBroker    BrokerResolver
	signalNotifier   signals.Notifier
	riskValidator    *risk.RiskValidator
	pairSvc          *pairs.Service
	mux              *http.ServeMux
}

func NewServer(cfg *config.Config, store *storage.PostgresStorage, subs *subscription.Service, tierSvc *tier.Service, paySvc *payments.Service, usersSvc *users.Service, resolve BrokerResolver, notifier signals.Notifier, validator *risk.RiskValidator, pairSvc *pairs.Service) *Server {
	s := &Server{
		cfg:              cfg,
		storage:          store,
		subs:             subs,
		tier:             tierSvc,
		payments:         paySvc,
		users:            usersSvc,
		resolveBroker:    resolve,
		signalNotifier:   notifier,
		riskValidator:    validator,
		pairSvc:          pairSvc,
		mux:              http.NewServeMux(),
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
	s.mux.HandleFunc("/api/v1/auth/telegram/webapp", s.handleTelegramWebAppAuth)
	s.mux.HandleFunc("/api/v1/auth/email", s.handleEmailLogin)
	s.mux.HandleFunc("/api/v1/auth/me", s.withUser(s.handleAuthMe))
	s.mux.HandleFunc("/api/v1/brokers/types", s.withUser(s.handleBrokerTypes))
	s.mux.HandleFunc("/api/v1/brokers/connection", s.withUser(s.handleBrokerConnection))
	s.mux.HandleFunc("/api/v1/brokers/test", s.withUser(s.handleBrokerTest))
	s.mux.HandleFunc("/api/v1/status", s.withUser(s.handleStatus))
	s.mux.HandleFunc("/api/v1/account", s.withUser(s.handleAccount))
	s.mux.HandleFunc("/api/v1/positions", s.withUser(s.handlePositions))
	s.mux.HandleFunc("/api/v1/trades", s.withUser(s.handleTrades))
	s.mux.HandleFunc("/api/v1/subscription", s.withUser(s.handleSubscription))
	s.mux.HandleFunc("/api/v1/tiers", s.handleTiers)
	s.mux.HandleFunc("/api/v1/miniapp/dashboard", s.withUser(s.handleMiniAppDashboard))
	s.mux.HandleFunc("/api/v1/payments/binance/order", s.withUser(s.handlePaymentOrderCreate))
	s.mux.HandleFunc("/api/v1/payments/binance/confirm", s.withUser(s.handlePaymentOrderConfirm))
	s.mux.HandleFunc("/api/v1/payments/binance/webhook", s.handleBinancePayWebhook)
	s.mux.HandleFunc("/api/v1/trading-pairs", s.withUser(s.handleTradingPairs))
	s.mux.HandleFunc("/api/v1/admin/stats", s.withAdmin(s.handleAdminStats))
	s.mux.HandleFunc("/api/v1/admin/trades", s.withAdmin(s.handleAdminTrades))
	s.mux.HandleFunc("/api/v1/admin/users", s.withAdmin(s.handleAdminUsers))
	s.mux.HandleFunc("/api/v1/admin/activate", s.withAdmin(s.handleAdminActivate))
	s.mux.HandleFunc("/api/v1/admin/users/block", s.withAdmin(s.handleAdminBlockUser))
	s.mux.HandleFunc("/api/v1/admin/users/revoke", s.withAdmin(s.handleAdminRevokeSubscription))
	s.mux.HandleFunc("/api/v1/admin/signals/broadcast", s.withAdmin(s.handleAdminBroadcastSignal))

	s.registerStatic()
}

func (s *Server) handleTradingPairs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleTradingPairsGet(w, r)
	case http.MethodPut:
		s.handleTradingPairsPut(w, r)
	default:
		methodNotAllowed(w)
	}
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
		// Required for Telegram Login popup (oauth.telegram.org postMessage).
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
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

func userIDFrom(r *http.Request) int64 {
	return r.Context().Value(userIDKey).(int64)
}
