package api

import (
	"context"
	"net/http"

	"forex-bot/internal/auth"
)

type ctxKey int

const (
	userIDKey ctxKey = 1
	roleKey   ctxKey = 2
)

func (s *Server) withUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "log in with Telegram or email")
			return
		}
		uid, err := auth.Verify(s.sessionSecret(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session — log in again")
			return
		}

		isAdmin := s.cfg.IsAdmin(uid)
		role := auth.ResolveRole(isAdmin)

		user, _ := s.storage.GetUserByTelegramID(uid)
		if user != nil && user.IsBlocked && !isAdmin {
			writeError(w, http.StatusForbidden, "account blocked — contact support")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, uid)
		ctx = context.WithValue(ctx, roleKey, role)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) withAdmin(next http.HandlerFunc) http.HandlerFunc {
	return s.withUser(func(w http.ResponseWriter, r *http.Request) {
		if roleFrom(r) != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	})
}

func (s *Server) withPermission(perm string, next http.HandlerFunc) http.HandlerFunc {
	return s.withUser(func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasPermission(roleFrom(r), perm) {
			writeError(w, http.StatusForbidden, "permission denied")
			return
		}
		next(w, r)
	})
}

func roleFrom(r *http.Request) auth.Role {
	if v, ok := r.Context().Value(roleKey).(auth.Role); ok {
		return v
	}
	return auth.RoleUser
}

func (s *Server) buildACLProfile(uid int64) auth.Profile {
	isAdmin := s.cfg.IsAdmin(uid)
	role := auth.ResolveRole(isAdmin)
	user, _ := s.storage.GetUserByTelegramID(uid)
	blocked := user != nil && user.IsBlocked
	canTrade, tradeMsg := s.subs.CanTrade(uid)
	if blocked && !isAdmin {
		canTrade = false
		tradeMsg = "account blocked"
	}
	return auth.Profile{
		TelegramID:   uid,
		Role:         role,
		IsAdmin:      isAdmin,
		Permissions:  auth.PermissionsFor(role),
		IsBlocked:    blocked,
		CanTrade:     canTrade,
		TradeMessage: tradeMsg,
	}
}
