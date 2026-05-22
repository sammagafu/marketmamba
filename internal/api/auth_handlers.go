package api

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
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
	token, err := s.completeTelegramLogin(telegramlogin.LoginData{
		ID: req.ID, FirstName: req.FirstName, LastName: req.LastName,
		Username: req.Username, PhotoURL: req.PhotoURL, AuthDate: req.AuthDate, Hash: req.Hash,
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user, _ := s.storage.GetUserByTelegramID(req.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_token": token,
		"telegram_id":   req.ID,
		"user":          user,
		"is_admin":      s.cfg.IsAdmin(req.ID),
	})
}

// Telegram Login Widget redirect (data-auth-url) — avoids iframe domain issues.
func (s *Server) handleTelegramLoginCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	q := r.URL.Query()
	id, _ := strconv.ParseInt(q.Get("id"), 10, 64)
	authDate, _ := strconv.ParseInt(q.Get("auth_date"), 10, 64)
	data := telegramlogin.LoginData{
		ID: id, FirstName: q.Get("first_name"), LastName: q.Get("last_name"),
		Username: q.Get("username"), PhotoURL: q.Get("photo_url"),
		AuthDate: authDate, Hash: q.Get("hash"),
	}
	token, err := s.completeTelegramLogin(data)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><body><p>Login failed: %s</p><p><a href="/">Back</a></p></body></html>`, html.EscapeString(err.Error()))
		return
	}
	home := s.cfg.App.PublicSiteURL + "/"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><body><script>
localStorage.setItem('mm_session', %q);
localStorage.setItem('mm_telegram_id', %q);
localStorage.removeItem('mm_api_key');
location.replace(%q);
</script></body></html>`, token, strconv.FormatInt(id, 10), home)
}

func (s *Server) handleTelegramOIDCLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	ou, err := telegramlogin.VerifyIDToken(s.cfg.Telegram.BotClientID, req.IDToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if _, err := s.users.RegisterFromLogin(ou.TelegramID, ou.Username, ou.FirstName, ou.LastName); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := auth.Issue(s.sessionSecret(), ou.TelegramID, 7*24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := s.storage.GetUserByTelegramID(ou.TelegramID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_token": token,
		"telegram_id":   ou.TelegramID,
		"user":          user,
		"is_admin":      s.cfg.IsAdmin(ou.TelegramID),
	})
}

func (s *Server) completeTelegramLogin(data telegramlogin.LoginData) (string, error) {
	if err := telegramlogin.Verify(s.cfg.Telegram.BotToken, data, 24*time.Hour); err != nil {
		return "", err
	}
	if _, err := s.users.RegisterFromLogin(data.ID, data.Username, data.FirstName, data.LastName); err != nil {
		return "", err
	}
	return auth.Issue(s.sessionSecret(), data.ID, 7*24*time.Hour)
}

type emailLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleEmailLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req emailLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	admin, err := s.storage.GetWebAdminByEmail(req.Email)
	if err != nil || admin == nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if !auth.CheckPassword(admin.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if !s.cfg.IsAdmin(admin.TelegramID) {
		writeError(w, http.StatusForbidden, "not configured as Telegram admin — add telegram id to TELEGRAM_ADMIN_USER_IDS")
		return
	}
	user, _ := s.storage.GetUserByTelegramID(admin.TelegramID)
	token, err := auth.Issue(s.sessionSecret(), admin.TelegramID, 7*24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_token": token,
		"telegram_id":   admin.TelegramID,
		"email":         admin.Email,
		"user":          user,
		"is_admin":      true,
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
