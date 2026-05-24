package api

import (
	"encoding/json"
	"net/http"

	"forex-bot/internal/models"
)

func (s *Server) handleTradingPairsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	if s.pairSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "pair preferences unavailable")
		return
	}
	resp, err := s.pairSvc.GetResponse(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleTradingPairsPut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	if s.pairSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "pair preferences unavailable")
		return
	}
	var req struct {
		Pairs       []models.UserTradingPair     `json:"pairs"`
		SignalTypes *models.SignalTypePreferences `json:"signal_types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.SignalTypes != nil {
		if err := s.pairSvc.SetSignalTypes(uid, *req.SignalTypes); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if len(req.Pairs) > 0 {
		if err := s.pairSvc.SetPreferences(uid, req.Pairs); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else if req.SignalTypes == nil {
		writeError(w, http.StatusBadRequest, "provide pairs or signal_types")
		return
	}
	resp, err := s.pairSvc.GetResponse(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
