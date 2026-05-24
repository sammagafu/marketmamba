package subscription

import (
	"fmt"
	"time"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/tier"
	"forex-bot/internal/utils"
)

type Service struct {
	store *storage.PostgresStorage
	cfg   *config.Config
	tier *tier.Service
}

func NewService(store *storage.PostgresStorage, cfg *config.Config) *Service {
	return &Service{store: store, cfg: cfg}
}

// SetTier wires tier quota checks into subscription status.
func (s *Service) SetTier(t *tier.Service) {
	s.tier = t
}

// CanAutoTrade checks subscription and optional admin approval for /autostart execution.
func (s *Service) CanAutoTrade(userID int64, isAdmin bool) (bool, string) {
	ok, msg := s.CanTrade(userID)
	if !ok {
		return false, msg
	}
	if !s.cfg.App.AutoTradeRequiresApproval || isAdmin {
		return true, ""
	}
	state, err := s.store.GetBotState(userID)
	if err != nil || state == nil {
		return false, "bot state not found — use /start first"
	}
	if !state.AutoTradeApproved {
		return false, "auto-trade pending admin approval — contact support or wait for /approveauto"
	}
	return true, ""
}

func (s *Service) CanTrade(userID int64) (bool, string) {
	if !s.cfg.App.SubscriptionRequired {
		return true, ""
	}
	sub, err := s.store.GetActiveSubscription(userID)
	if err != nil {
		return false, "could not verify subscription"
	}
	if sub == nil {
		return false, s.cfg.App.SubscriptionContactMessage
	}
	if sub.Status != "active" {
		return false, "subscription is not active — contact support or use /subscribe"
	}
	if sub.ExpiresAt != nil && sub.ExpiresAt.Before(time.Now()) {
		return false, "subscription expired — use /subscribe for renewal"
	}
	return true, ""
}

func (s *Service) EnsureTrial(userID int64) error {
	existing, err := s.store.GetActiveSubscription(userID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	exp := time.Now().AddDate(0, 0, s.cfg.App.FreeTrialDays)
	sub := &models.Subscription{
		ID:          utils.GenerateID("sub"),
		UserID:      userID,
		Plan:        "trial",
		Status:      "active",
		ExpiresAt:   &exp,
		Notes:       fmt.Sprintf("Free trial — %d days", s.cfg.App.FreeTrialDays),
		ActivatedBy: "system",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.store.CreateSubscription(sub)
}

func (s *Service) ActivateManual(userID int64, days int, plan, notes string, adminID int64) (*models.Subscription, error) {
	if days <= 0 {
		days = 30
	}
	if plan == "" {
		plan = "manual"
	}
	exp := time.Now().AddDate(0, 0, days)
	sub := &models.Subscription{
		ID:          utils.GenerateID("sub"),
		UserID:      userID,
		Plan:        plan,
		Status:      "active",
		ExpiresAt:   &exp,
		Notes:       notes,
		ActivatedBy: fmt.Sprintf("admin:%d", adminID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.store.DeactivateSubscriptions(userID); err != nil {
		return nil, err
	}
	if err := s.store.CreateSubscription(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// ActivatePaid extends or creates a paid monthly plan after USDT payment.
func (s *Service) ActivatePaid(userID int64, days int, plan, notes string) (*models.Subscription, error) {
	if days <= 0 {
		days = 30
	}
	if plan == "" {
		plan = "monthly"
	}
	existing, _ := s.store.GetActiveSubscription(userID)
	start := time.Now()
	if existing != nil && existing.ExpiresAt != nil && existing.ExpiresAt.After(start) {
		start = *existing.ExpiresAt
	}
	exp := start.AddDate(0, 0, days)
	sub := &models.Subscription{
		ID:          utils.GenerateID("sub"),
		UserID:      userID,
		Plan:        plan,
		Status:      "active",
		ExpiresAt:   &exp,
		Notes:       notes,
		ActivatedBy: "binance_usdt",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.store.DeactivateSubscriptions(userID); err != nil {
		return nil, err
	}
	if err := s.store.CreateSubscription(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// SubscriptionStatus returns trial/paid state for UI.
func (s *Service) SubscriptionStatus(userID int64) map[string]interface{} {
	sub, _ := s.store.GetActiveSubscription(userID)
	canTrade, msg := s.CanTrade(userID)
	out := map[string]interface{}{
		"can_trade":             canTrade,
		"message":               msg,
		"subscription_required": s.cfg.App.SubscriptionRequired,
		"trial_days":            s.cfg.App.FreeTrialDays,
		"price_usdt":            s.cfg.Payments.SubscriptionPriceUSDT,
	}
	if sub != nil {
		out["subscription"] = sub
		out["plan"] = sub.Plan
		out["status"] = sub.Status
		if sub.ExpiresAt != nil {
			out["expires_at"] = sub.ExpiresAt.Format(time.RFC3339)
			out["days_left"] = int(time.Until(*sub.ExpiresAt).Hours() / 24)
		}
	}
	if s.tier != nil {
		if snap, err := s.tier.Snapshot(userID); err == nil {
			out["tier"] = snap
		}
	}
	return out
}

func (s *Service) GetForUser(userID int64) (*models.Subscription, error) {
	return s.store.GetActiveSubscription(userID)
}
