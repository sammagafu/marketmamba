package payments

import (
	"fmt"
	"strings"
	"time"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
	"forex-bot/internal/utils"
)

type Service struct {
	store *storage.PostgresStorage
	subs  *subscription.Service
	cfg   *config.Config
	binance *BinancePayClient
}

func NewService(store *storage.PostgresStorage, subs *subscription.Service, cfg *config.Config) *Service {
	s := &Service{store: store, subs: subs, cfg: cfg}
	if cfg.Payments.BinancePayAPIKey != "" {
		s.binance = &BinancePayClient{
			APIKey:    cfg.Payments.BinancePayAPIKey,
			SecretKey: cfg.Payments.BinancePaySecret,
			CertSN:    cfg.Payments.BinancePayCertSN,
		}
	}
	return s
}

func (s *Service) Pricing() map[string]interface{} {
	return map[string]interface{}{
		"trial_days":           s.cfg.App.FreeTrialDays,
		"price_usdt":           s.cfg.Payments.SubscriptionPriceUSDT,
		"billing_period_days":  s.cfg.Payments.SubscriptionDays,
		"currency":             "USDT",
		"binance_pay_enabled":  s.binance != nil && s.binance.Enabled(),
		"binance_uid":          s.cfg.Payments.BinanceUID,
		"binance_network":      s.cfg.Payments.BinanceNetwork,
	}
}

// CreateMonthlyOrder starts a 10 USDT monthly subscription payment.
func (s *Service) CreateMonthlyOrder(userID int64) (*models.PaymentOrder, error) {
	_ = s.store.ExpireStalePaymentOrders()
	amount := s.cfg.Payments.SubscriptionPriceUSDT
	if amount <= 0 {
		amount = 10
	}
	now := time.Now()
	tradeNo := fmt.Sprintf("MM%d%d", userID, now.Unix())
	order := &models.PaymentOrder{
		ID:              utils.GenerateID("pay"),
		UserID:          userID,
		AmountUSDT:      amount,
		Plan:            "monthly",
		Status:          "pending",
		MerchantTradeNo: tradeNo,
		PayMethod:       "binance_manual",
		ExpiresAt:       now.Add(2 * time.Hour),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if s.binance != nil && s.binance.Enabled() {
		prepay, url, err := s.binance.CreateUSDTOrder(
			tradeNo, amount, "Market Mamba — 1 month",
		)
		if err != nil {
			return nil, err
		}
		order.PayMethod = "binance_pay"
		order.BinancePrepayID = prepay
		order.CheckoutURL = url
	} else if s.cfg.Payments.BinanceUID != "" {
		order.PayMethod = "binance_transfer"
		order.CheckoutURL = fmt.Sprintf(
			"binance://pay/send?asset=USDT&amount=%.2f&uid=%s&memo=%s",
			amount, s.cfg.Payments.BinanceUID, tradeNo,
		)
	} else {
		return nil, fmt.Errorf("binance payment not configured — set BINANCE_PAY_* or BINANCE_PAY_UID on server")
	}

	if err := s.store.CreatePaymentOrder(order); err != nil {
		return nil, err
	}
	return order, nil
}

// ConfirmOrder marks paid and extends subscription (webhook or admin).
func (s *Service) ConfirmOrder(order *models.PaymentOrder, txRef string) error {
	if order == nil || order.Status == "paid" {
		return nil
	}
	now := time.Now()
	order.Status = "paid"
	order.TxReference = txRef
	order.PaidAt = &now
	order.UpdatedAt = now
	if err := s.store.UpdatePaymentOrder(order); err != nil {
		return err
	}
	days := s.cfg.Payments.SubscriptionDays
	if days <= 0 {
		days = 30
	}
	_, err := s.subs.ActivatePaid(order.UserID, days, "monthly", fmt.Sprintf("Binance USDT %.2f (%s)", order.AmountUSDT, order.MerchantTradeNo))
	return err
}

// SubmitTxReference lets user submit Binance tx id after manual transfer.
func (s *Service) SubmitTxReference(userID int64, orderID, txRef string) (*models.PaymentOrder, error) {
	txRef = strings.TrimSpace(txRef)
	if txRef == "" {
		return nil, fmt.Errorf("transaction reference required")
	}
	order, err := s.store.GetPaymentOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil || order.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != "pending" {
		return nil, fmt.Errorf("order is %s", order.Status)
	}
	if time.Now().After(order.ExpiresAt) {
		order.Status = "expired"
		_ = s.store.UpdatePaymentOrder(order)
		return nil, fmt.Errorf("order expired — create a new payment")
	}
	order.TxReference = txRef
	order.Status = "confirming"
	order.UpdatedAt = time.Now()
	if err := s.store.UpdatePaymentOrder(order); err != nil {
		return nil, err
	}
	// Auto-activate when user submits ref (Binance Pay webhook is authoritative when configured).
	if err := s.ConfirmOrder(order, txRef); err != nil {
		return nil, err
	}
	return order, nil
}

// HandleBinanceWebhook processes Binance Pay notification (simplified).
func (s *Service) HandleBinanceWebhook(merchantTradeNo, bizStatus string) error {
	if strings.ToUpper(bizStatus) != "PAY_SUCCESS" {
		return nil
	}
	order, err := s.store.GetPaymentOrderByMerchantTradeNo(merchantTradeNo)
	if err != nil || order == nil {
		return fmt.Errorf("order not found")
	}
	return s.ConfirmOrder(order, "binance_webhook")
}
