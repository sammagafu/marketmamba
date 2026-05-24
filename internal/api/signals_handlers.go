package api

import (
	"encoding/json"
	"net/http"
	"time"

	"forex-bot/internal/models"
	"forex-bot/internal/signals"
)

func (s *Server) handleAdminBroadcastSignal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		Symbol     string  `json:"symbol"`
		Type       string  `json:"type"`
		StopLoss   float64 `json:"stop_loss"`
		TakeProfit float64 `json:"take_profit"`
		Strength   float64 `json:"strength"`
		Generate   bool    `json:"generate"`
		Force      bool    `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if s.signalNotifier == nil || s.riskValidator == nil {
		writeError(w, http.StatusServiceUnavailable, "signal notifier not configured")
		return
	}

	symbol := req.Symbol
	if symbol == "" {
		symbol = s.cfg.App.SignalBroadcastSymbol
	}
	minStrength := s.cfg.App.SignalMinStrength

	var signal *models.TradeSignal
	var qualErr error

	if req.Generate || (req.Symbol == "" && req.Type == "") {
		if symbol == "" {
			for _, sym := range s.cfg.SignalSymbols() {
				signal, qualErr = signals.GenerateQualified(sym, minStrength, 0, s.riskValidator)
				if qualErr == nil {
					break
				}
			}
		} else {
			signal, qualErr = signals.GenerateQualified(symbol, minStrength, 0, s.riskValidator)
		}
		if qualErr != nil && signal == nil && !req.Force {
			writeError(w, http.StatusBadRequest, qualErr.Error())
			return
		}
	} else {
		if req.Type != "BUY" && req.Type != "SELL" {
			writeError(w, http.StatusBadRequest, "type must be BUY or SELL")
			return
		}
		signal = &models.TradeSignal{
			Symbol:          symbol,
			Type:            req.Type,
			StopLoss:        req.StopLoss,
			TakeProfit:      req.TakeProfit,
			Strength:        req.Strength,
			RiskRewardRatio: 2.0,
			TriggeredAt:     time.Now(),
		}
		if signal.Strength <= 0 {
			signal.Strength = 0.85
		}
	}

	if signal == nil {
		writeError(w, http.StatusBadRequest, "no signal meets requirements — try generate:true or check filters")
		return
	}

	n, err := signals.PublishManual(s.storage, s.subs, s.tier, s.pairSvc, s.signalNotifier, s.riskValidator, minStrength, signal, req.Force)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if n == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"sent":    0,
			"message": "signal qualified but no eligible subscribers",
			"signal":  signal,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sent":   n,
		"signal": signal,
	})
}
