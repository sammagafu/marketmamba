package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"forex-bot/internal/auth"
	"forex-bot/internal/models"
	"forex-bot/internal/positions"
	"forex-bot/internal/telegramlogin"
	"forex-bot/internal/tier"
)

type webappAuthRequest struct {
	InitData string `json:"init_data"`
}

func (s *Server) handleTelegramWebAppAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req webappAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	wu, err := telegramlogin.VerifyWebAppInitData(s.cfg.Telegram.BotToken, req.InitData, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if _, err := s.users.RegisterFromLogin(wu.ID, wu.Username, wu.FirstName, wu.LastName); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := auth.Issue(s.sessionSecret(), wu.ID, s.cfg.SessionTTL())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := s.storage.GetUserByTelegramID(wu.ID)
	writeJSON(w, http.StatusOK, s.loginResponse(token, wu.ID, user, ""))
}

func (s *Server) handleMiniAppDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	trades, _ := s.storage.ListTradesByUser(uid, 100)
	if trades == nil {
		trades = []*models.Trade{}
	}
	orders, _ := s.storage.ListPaymentOrdersByUser(uid, 10)
	if orders == nil {
		orders = []*models.PaymentOrder{}
	}
	today, _ := s.storage.GetDailyStats(uid, time.Now())
	subStatus := s.subs.SubscriptionStatus(uid)

	posList := []*models.Position{}
	if b, err := s.resolveBroker(uid); err == nil {
		if userPos, err := positions.ListOpenForUser(s.storage, uid, b); err == nil && userPos != nil {
			posList = userPos
		}
	} else if userPos, err := positions.ListOpenForUser(s.storage, uid, nil); err == nil && userPos != nil {
		posList = userPos
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"trades":                trades,
		"positions":             posList,
		"payments":              orders,
		"daily_stats":           today,
		"subscription":          subStatus,
		"packages":              tier.PublicPackages(s.cfg.Payments.SubscriptionPriceUSDT, s.cfg.App.FreeTrialDays),
		"pricing":               s.payments.Pricing(),
		"public_site_url":       s.cfg.App.PublicSiteURL,
		"connect_url":           strings.TrimRight(s.cfg.App.PublicSiteURL, "/") + "/#/connect",
		"value_proposition":     s.cfg.App.ValueProposition,
		"contact_us_url":        s.cfg.App.ContactUsURL,
		"contact_us_label":      s.cfg.App.ContactUsLabel,
		"payment_note":          "USDT via Binance only — no cards.",
		"telegram_bot_username": s.cfg.Telegram.BotUsername,
	})
}
