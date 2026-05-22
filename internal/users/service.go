package users

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
	"forex-bot/internal/pairs"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
	"forex-bot/internal/utils"
)

type Service struct {
	store   *storage.PostgresStorage
	subs    *subscription.Service
	pairSvc *pairs.Service
	cfg     *config.Config
}

func NewService(store *storage.PostgresStorage, subs *subscription.Service, cfg *config.Config) *Service {
	return &Service{
		store:   store,
		subs:    subs,
		pairSvc: pairs.NewService(store, cfg),
		cfg:     cfg,
	}
}

func (s *Service) RegisterFromTelegram(from *tgbotapi.User) (*models.User, error) {
	now := time.Now()
	u := &models.User{
		ID:         utils.GenerateID("usr"),
		TelegramID: from.ID,
		Username:   from.UserName,
		FirstName:  from.FirstName,
		LastName:   from.LastName,
		CreatedAt:  now,
		LastSeenAt: now,
	}
	if err := s.store.UpsertUser(u); err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(from.ID); err != nil {
		return nil, err
	}
	if err := s.subs.EnsureTrial(from.ID); err != nil {
		return nil, err
	}
	return s.store.GetUserByTelegramID(from.ID)
}

func (s *Service) Touch(telegramID int64) error {
	return s.store.UpdateUserLastSeen(telegramID, time.Now())
}

// RegisterFromLogin creates or updates a user from Telegram Login Widget data.
func (s *Service) RegisterFromLogin(telegramID int64, username, firstName, lastName string) (*models.User, error) {
	now := time.Now()
	u := &models.User{
		ID:         utils.GenerateID("usr"),
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		CreatedAt:  now,
		LastSeenAt: now,
	}
	if err := s.store.UpsertUser(u); err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(telegramID); err != nil {
		return nil, err
	}
	if err := s.subs.EnsureTrial(telegramID); err != nil {
		return nil, err
	}
	return s.store.GetUserByTelegramID(telegramID)
}

func (s *Service) ensureDefaults(userID int64) error {
	if _, err := s.store.GetBotState(userID); err != nil {
		state := &models.BotState{
			ID:                  utils.GenerateID("state"),
			UserID:              userID,
			IsPaused:            false,
			AutoTradingActive:   false,
			AutoTradeApproved:   s.cfg.IsAdmin(userID),
			DailyLossHit:        false,
			LastActiveAt:        time.Now(),
			UpdatedAt:           time.Now(),
		}
		if err := s.store.CreateBotState(state); err != nil {
			return err
		}
	}
	if _, err := s.store.GetRiskSettings(userID); err != nil {
		rs := &models.RiskSettings{
			ID:              utils.GenerateID("risk"),
			UserID:          userID,
			MaxRiskPerTrade: 0.005,
			MaxDailyLoss:    0.02,
			MaxOpenTrades:   2,
			MaxTradesPerDay: 10,
			RiskRewardRatio: 1.0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := s.store.CreateRiskSettings(rs); err != nil {
			return err
		}
	}
	if has, _ := s.store.HasUserTradingPairs(userID); !has && s.pairSvc != nil {
		if err := s.pairSvc.SeedDefaults(userID); err != nil {
			return err
		}
	}
	return nil
}
