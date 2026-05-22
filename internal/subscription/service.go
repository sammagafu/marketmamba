package subscription

import (
	"fmt"
	"time"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/utils"
)

type Service struct {
	store  *storage.PostgresStorage
	cfg    *config.Config
}

func NewService(store *storage.PostgresStorage, cfg *config.Config) *Service {
	return &Service{store: store, cfg: cfg}
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
		Notes:       "Auto trial — testing period",
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

func (s *Service) GetForUser(userID int64) (*models.Subscription, error) {
	return s.store.GetActiveSubscription(userID)
}
