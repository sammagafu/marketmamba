package api

import (
	"encoding/json"
	"net/http"

	"forex-bot/internal/accounts"
	"forex-bot/internal/broker"
	"forex-bot/internal/models"
	"forex-bot/internal/positions"
	"forex-bot/internal/tier"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	payload := map[string]interface{}{
		"status":  "ok",
		"service": "market-mamba",
		"app_env": s.cfg.App.Environment,
	}
	if err := s.storage.Health(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "degraded",
			"error":  "database",
		})
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handlePublicConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	payload := map[string]interface{}{
		"app":                   "Market Mamba",
		"public_mode":           s.cfg.App.PublicMode,
		"subscription_required": s.cfg.App.SubscriptionRequired,
		"subscription_message":  s.cfg.App.SubscriptionContactMessage,
		"free_trial_days":       s.cfg.App.FreeTrialDays,
		"telegram_bot_username":  s.cfg.Telegram.BotUsername,
		"telegram_client_id":     s.cfg.Telegram.BotClientID,
		"telegram_login_domain":  s.cfg.Telegram.LoginDomain,
		"public_site_url":        s.cfg.App.PublicSiteURL,
		"telegram_login_enabled": s.cfg.Telegram.BotToken != "",
		"mini_app_url":           s.cfg.Payments.MiniAppURL,
		"session_ttl_days":       s.cfg.App.WebSessionTTLDays,
		"subscription_price_usdt": s.cfg.Payments.SubscriptionPriceUSDT,
		"subscription_days":      s.cfg.Payments.SubscriptionDays,
		"trial_days":             s.cfg.App.FreeTrialDays,
		"binance_pay_enabled":    s.payments != nil && s.cfg.Payments.BinancePayAPIKey != "",
		"signal_broadcast":       s.cfg.App.SignalBroadcastEnabled,
		"signal_symbols":         s.cfg.SignalSymbols(),
	}
	if stats, err := s.storage.GetUserStats(); err == nil && stats != nil {
		payload["total_trades"] = stats.TotalTrades
		payload["total_users"] = stats.TotalUsers
		payload["open_trades"] = stats.OpenTrades
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleBrokerTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"brokers":                broker.SupportedBrokerTypes(),
		"brands":                 broker.SupportedBrands(),
		"metaapi_brands":         broker.MetaAPIBrands(),
		"metaapi_shared_token":   broker.UsesSharedMetaAPIToken(),
		"connection_method":      "MT brokers (Deriv, Exness, Tickmill, etc.) use the MetaAPI cloud bridge to your MT4/MT5 login.",
	})
}

func (s *Server) getBrokerConnection(w http.ResponseWriter, r *http.Request) {
	uid := userIDFrom(r)
	conns, err := s.storage.ListBrokerConnections(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	primary, _ := s.storage.GetPrimaryBrokerConnection(uid)
	payload := map[string]interface{}{
		"connection":  nil,
		"connections": []interface{}{},
	}
	if primary != nil {
		payload["connection"] = map[string]interface{}{
			"id":          primary.ID,
			"provider":    primary.Provider,
			"label":       primary.Label,
			"is_primary":  primary.IsPrimary,
			"updated_at":  primary.UpdatedAt,
		}
	}
	list := make([]map[string]interface{}, 0, len(conns))
	for _, c := range conns {
		list = append(list, map[string]interface{}{
			"id":         c.ID,
			"provider":   c.Provider,
			"label":      c.Label,
			"is_primary": c.IsPrimary,
			"updated_at": c.UpdatedAt,
		})
	}
	payload["connections"] = list
	writeJSON(w, http.StatusOK, payload)
}

type saveBrokerRequest struct {
	BrandID     string            `json:"brand_id"`
	Provider    string            `json:"provider"`
	Label       string            `json:"label"`
	Credentials map[string]string `json:"credentials"`
}

func (s *Server) saveBrokerConnection(w http.ResponseWriter, r *http.Request) {
	uid := userIDFrom(r)
	var req saveBrokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if s.tier != nil {
		if err := s.tier.CanAddBroker(uid); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	var saveErr error
	if req.BrandID != "" {
		saveErr = broker.SaveBrandConnection(s.storage, s.cfg.App.BrokerEncryptionKey, uid, req.BrandID, req.Label, req.Credentials)
	} else {
		saveErr = broker.SaveConnection(s.storage, s.cfg.App.BrokerEncryptionKey, uid, req.Provider, req.Label, req.Credentials)
	}
	if saveErr != nil {
		writeError(w, http.StatusBadRequest, saveErr.Error())
		return
	}
	if _, err := broker.ResolveBrokerAndSync(s.storage, uid, s.cfg.App.BrokerEncryptionKey, s.cfg.Broker.Provider); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	provider := req.Provider
	if req.BrandID != "" {
		if b, ok := broker.BrandByID(req.BrandID); ok {
			provider = b.AdapterID
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "saved", "provider": provider, "brand_id": req.BrandID})
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

func (s *Server) handleBrokerTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req saveBrokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	provider := req.Provider
	creds := req.Credentials
	if req.BrandID != "" {
		var err error
		provider, creds, _, err = broker.ResolveBrandConnection(req.BrandID, req.Label, req.Credentials)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		creds = broker.ApplySharedMetaAPIToken(creds)
	}
	b, err := broker.NewFromProvider(provider, creds)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bal, err := b.GetBalance()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	uid := userIDFrom(r)
	if err := accounts.SyncFromBroker(s.storage, uid, provider, b); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "balance": bal})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	state, _ := s.storage.GetBotState(uid)
	conn, _ := s.storage.GetActiveBrokerConnection(uid)
	provider := s.cfg.Broker.Provider
	if conn != nil {
		provider = conn.Provider
	}
	canTrade, tradeMsg := s.subs.CanTrade(uid)
	resp := map[string]interface{}{
		"app": "Market Mamba", "env": s.cfg.App.Environment, "provider": provider,
		"can_trade": canTrade, "trade_message": tradeMsg,
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
	uid := userIDFrom(r)
	if ok, msg := s.subs.CanTrade(uid); !ok {
		writeError(w, http.StatusForbidden, msg)
		return
	}
	b, err := s.resolveBroker(uid)
	if err != nil {
		writeError(w, http.StatusBadRequest, "connect a broker first: "+err.Error())
		return
	}
	if conn, _ := s.storage.GetActiveBrokerConnection(uid); conn != nil {
		_ = accounts.SyncFromBroker(s.storage, uid, conn.Provider, b)
	}
	bal, _ := b.GetBalance()
	eq, _ := b.GetEquity()
	writeJSON(w, http.StatusOK, map[string]interface{}{"balance": bal, "equity": eq})
}

func (s *Server) handleTrades(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	trades, err := s.storage.ListTradesByUser(uid, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if trades == nil {
		trades = []*models.Trade{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"trades": trades})
}

func (s *Server) handlePositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	b, err := s.resolveBroker(uid)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"positions": []interface{}{}})
		return
	}
	userPos, err := positions.ListOpenForUser(s.storage, uid, b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if userPos == nil {
		userPos = []*models.Position{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"positions": userPos})
}

func (s *Server) handleSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	writeJSON(w, http.StatusOK, s.subs.SubscriptionStatus(uid))
}

func (s *Server) handleTiers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plans": tier.AllPlans(),
	})
}

func (s *Server) handleAdminTrades(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	trades, err := s.storage.ListRecentTrades(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if trades == nil {
		trades = []*models.Trade{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"trades": trades})
}

func (s *Server) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	stats, err := s.storage.GetUserStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload := map[string]interface{}{
		"stats": stats,
	}
	if counts, err := s.storage.CountBrokerConnectionsByProvider(); err == nil {
		payload["broker_connections"] = counts
	}
	brands := make([]map[string]string, 0)
	for _, b := range broker.SupportedBrands() {
		brands = append(brands, map[string]string{
			"id": b.ID, "name": b.DisplayName, "adapter": b.AdapterID, "status": b.Status,
		})
	}
	payload["enabled_broker_brands"] = brands
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	users, err := s.storage.ListRecentUsers(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"users": users})
}

func (s *Server) handleAdminActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	adminID := userIDFrom(r)
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Days       int    `json:"days"`
		Plan       string `json:"plan"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	sub, err := s.subs.ActivateManual(req.TelegramID, req.Days, req.Plan, req.Notes, adminID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"subscription": sub})
}

func (s *Server) handleAdminBlockUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		TelegramID int64 `json:"telegram_id"`
		Blocked    bool  `json:"blocked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := s.storage.SetUserBlocked(req.TelegramID, req.Blocked); err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"telegram_id": req.TelegramID, "blocked": req.Blocked})
}

func (s *Server) handleAdminRevokeSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		TelegramID int64 `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := s.storage.RevokeActiveSubscription(req.TelegramID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"telegram_id": req.TelegramID, "revoked": true})
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
