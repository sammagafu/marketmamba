package api

import (
	"net/http"
	"strings"
)

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.App.WebAPIKey == "" {
			next(w, r)
			return
		}
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if key != s.cfg.App.WebAPIKey {
			writeError(w, http.StatusUnauthorized, "invalid or missing API key")
			return
		}
		next(w, r)
	}
}
