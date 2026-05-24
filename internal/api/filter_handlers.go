package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"forex-bot/internal/filter"
)

func (s *Server) handleFilterCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"filters": filter.Catalog(),
		"layers": []map[string]string{
			{"id": "market", "title": "Market quality"},
			{"id": "technical", "title": "Technical filters"},
			{"id": "setup", "title": "Setup & signal"},
			{"id": "risk", "title": "Risk envelope"},
			{"id": "platform", "title": "Platform policy"},
		},
	})
}

func (s *Server) handleFilterReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	if s.filters == nil {
		writeError(w, http.StatusServiceUnavailable, "filter service not configured")
		return
	}
	symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	if symbol == "" {
		symbol = s.cfg.App.SignalBroadcastSymbol
		if symbol == "" {
			symbol = "EURUSD"
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	report, err := s.filters.Report(ctx, symbol)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"report": report,
		"symbol": symbol,
	})
}
