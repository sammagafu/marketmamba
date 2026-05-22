package api

import (
	"encoding/json"
	"net/http"
	"time"

	"forex-bot/internal/auth"
	"forex-bot/internal/telegramlogin"
)

type telegramLoginRequest struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

func (s *Server) handleTelegramLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req telegramLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	data := telegramlogin.LoginData{
		ID: req.ID, FirstName: req.FirstName, LastName: req.LastName,
		Username: req.Username, PhotoURL: req.PhotoURL, AuthDate: req.AuthDate, Hash: req.Hash,
	}
	if err := telegramlogin.Verify(s.cfg.Telegram.BotToken, data, 24*time.Hour); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user, err := s.users.RegisterFromLogin(req.ID, req.Username, req.FirstName, req.LastName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := auth.Issue(s.sessionSecret(), req.ID, 7*24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_token": token,
		"telegram_id":   req.ID,
		"user":          user,
		"is_admin":      s.cfg.IsAdmin(req.ID),
	})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	user, _ := s.storage.GetUserByTelegramID(uid)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"telegram_id": uid,
		"user":        user,
		"is_admin":    s.cfg.IsAdmin(uid),
	})
}

func (s *Server) sessionSecret() string {
	if s.cfg.App.WebSessionSecret != "" {
		return s.cfg.App.WebSessionSecret
	}
	return s.cfg.App.BrokerEncryptionKey
}
