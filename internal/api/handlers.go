package api

import (
	"encoding/json"
	"net/http"
	"time"

	"forex-bot/internal/broker"
	"forex-bot/internal/models"
	"forex-bot/internal/secrets"
	"forex-bot/internal/utils"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleBrokerTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"brokers": broker.SupportedBrokerTypes(),
	})
}

func (s *Server) handleBrokerConnection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getBrokerConnection(w, r)
	case http.MethodPost:
		s.saveBrokerConnection(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) getBrokerConnection(w http.ResponseWriter, _ *http.Request) {
	conn, err := s.storage.GetActiveBrokerConnection(s.userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if conn == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"connection": nil})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"connection": map[string]interface{}{
			"provider":  conn.Provider,
			"label":     conn.Label,
			"is_active": conn.IsActive,
			"updated_at": conn.UpdatedAt,
		},
	})
}

type saveBrokerRequest struct {
	Provider    string            `json:"provider"`
	Label       string            `json:"label"`
	Credentials map[string]string `json:"credentials"`
}

func (s *Server) saveBrokerConnection(w http.ResponseWriter, r *http.Request) {
	var req saveBrokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}
	if !broker.IsLiveProvider(req.Provider) {
		writeError(w, http.StatusBadRequest, "this broker is not available yet; choose Mock (Demo) for now")
		return
	}
	if _, err := broker.NewFromProvider(req.Provider, req.Credentials); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	enc, err := secrets.EncryptJSON(s.cfg.App.BrokerEncryptionKey, req.Credentials)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	now := time.Now()
	conn := &models.BrokerConnection{
		ID:             utils.GenerateID("broker"),
		UserID:         s.userID,
		Provider:       req.Provider,
		Label:          req.Label,
		CredentialsEnc: enc,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.storage.UpsertBrokerConnection(conn); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.broker, _ = broker.NewFromProvider(req.Provider, req.Credentials)
	writeJSON(w, http.StatusOK, map[string]string{"message": "broker connection saved", "provider": req.Provider})
}

func (s *Server) handleBrokerTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req saveBrokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	b, err := broker.NewFromProvider(req.Provider, req.Credentials)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bal, err := b.GetBalance()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"balance": bal,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	state, _ := s.storage.GetBotState(s.userID)
	conn, _ := s.storage.GetActiveBrokerConnection(s.userID)
	provider := s.cfg.Broker.Provider
	if conn != nil {
		provider = conn.Provider
	}
	resp := map[string]interface{}{
		"app":      "Market Mamba",
		"env":      s.cfg.App.Environment,
		"provider": provider,
	}
	if state != nil {
		resp["is_paused"] = state.IsPaused
		resp["auto_trading"] = state.AutoTradingActive
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	bal, err := s.broker.GetBalance()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	eq, _ := s.broker.GetEquity()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"balance": bal,
		"equity":  eq,
	})
}

func (s *Server) handlePositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	positions, err := s.broker.GetOpenPositions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if positions == nil {
		positions = []*models.Position{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"positions": positions})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
