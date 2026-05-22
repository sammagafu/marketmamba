package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
)

func (s *Server) handlePaymentOrderCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	uid := userIDFrom(r)
	order, err := s.payments.CreateMonthlyOrder(uid)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order":        order,
		"pricing":      s.payments.Pricing(),
		"instructions": buildPaymentInstructions(s.cfg, order),
	})
}

func (s *Server) handlePaymentOrderConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		OrderID     string `json:"order_id"`
		TxReference string `json:"tx_reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	uid := userIDFrom(r)
	order, err := s.payments.SubmitTxReference(uid, req.OrderID, req.TxReference)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order":        order,
		"subscription": s.subs.SubscriptionStatus(uid),
	})
}

func (s *Server) handleBinancePayWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var data struct {
		MerchantTradeNo string `json:"merchantTradeNo"`
		BizStatus       string `json:"bizStatus"`
	}
	_ = json.Unmarshal(body, &data)
	if data.MerchantTradeNo == "" {
		var wrap struct {
			Data string `json:"data"`
		}
		_ = json.Unmarshal(body, &wrap)
		_ = json.Unmarshal([]byte(wrap.Data), &data)
	}
	if err := s.payments.HandleBinanceWebhook(data.MerchantTradeNo, data.BizStatus); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"returnCode": "SUCCESS"})
}

func buildPaymentInstructions(cfg *config.Config, order *models.PaymentOrder) map[string]string {
	out := map[string]string{
		"amount_usdt": fmt.Sprintf("%.2f", order.AmountUSDT),
		"reference":   order.MerchantTradeNo,
		"plan":        "1 month access",
	}
	if order.CheckoutURL != "" && order.PayMethod == "binance_pay" {
		out["step1"] = "Tap Open Binance Pay below"
		out["checkout_url"] = order.CheckoutURL
	}
	if cfg.Payments.BinanceUID != "" {
		out["binance_uid"] = cfg.Payments.BinanceUID
		out["network"] = cfg.Payments.BinanceNetwork
		out["memo"] = order.MerchantTradeNo
		out["step1"] = fmt.Sprintf("Send %.2f USDT to Binance UID %s (%s)", order.AmountUSDT, cfg.Payments.BinanceUID, cfg.Payments.BinanceNetwork)
		out["step2"] = "Put reference in memo: " + order.MerchantTradeNo
		out["step3"] = "Submit your transaction ID in the app"
	}
	return out
}
