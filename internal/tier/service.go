package tier

import (
	"fmt"
	"strings"
	"time"

	"forex-bot/internal/config"
	"forex-bot/internal/storage"
)

type Service struct {
	store *storage.PostgresStorage
	cfg   *config.Config
}

func NewService(store *storage.PostgresStorage, cfg *config.Config) *Service {
	return &Service{store: store, cfg: cfg}
}

func (s *Service) IsAdmin(userID int64) bool {
	return s.cfg != nil && s.cfg.IsAdmin(userID)
}

func (s *Service) planForUser(userID int64) string {
	if s.IsAdmin(userID) {
		return "manual"
	}
	sub, err := s.store.GetActiveSubscription(userID)
	if err != nil || sub == nil {
		return "trial"
	}
	return strings.ToLower(strings.TrimSpace(sub.Plan))
}

func (s *Service) Snapshot(userID int64) (Snapshot, error) {
	plan := s.planForUser(userID)
	limits := ForPlan(plan)
	now := time.Now()
	signals, longT, shortT, period, err := s.store.GetOrCreateUsage(userID, now)
	if err != nil {
		return Snapshot{}, err
	}
	brokers, err := s.store.CountActiveBrokerConnections(userID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Limits: limits,
		Usage: Usage{
			PeriodStart:     period.Format("2006-01-02"),
			SignalsReceived: signals,
			LongTrades:      longT,
			ShortTrades:     shortT,
			BrokerAccounts:  brokers,
		},
	}, nil
}

func (s *Service) CanAddBroker(userID int64) error {
	if s.IsAdmin(userID) {
		return nil
	}
	snap, err := s.Snapshot(userID)
	if err != nil {
		return err
	}
	if snap.Usage.BrokerAccounts >= snap.Limits.MaxBrokerAccounts {
		return fmt.Errorf(
			"broker account limit reached (%d/%d) — upgrade plan or remove a connection",
			snap.Usage.BrokerAccounts, snap.Limits.MaxBrokerAccounts,
		)
	}
	return nil
}

func (s *Service) CanReceiveSignal(userID int64) (bool, string) {
	if s.IsAdmin(userID) {
		return true, ""
	}
	snap, err := s.Snapshot(userID)
	if err != nil {
		return false, "could not verify plan usage"
	}
	if snap.Usage.SignalsReceived >= snap.Limits.MaxSignalsPerPeriod {
		return false, fmt.Sprintf(
			"signal limit reached (%d/%d this month) — upgrade your plan",
			snap.Usage.SignalsReceived, snap.Limits.MaxSignalsPerPeriod,
		)
	}
	return true, ""
}

func (s *Service) RecordSignal(userID int64) error {
	if s.IsAdmin(userID) {
		return nil
	}
	ok, msg := s.CanReceiveSignal(userID)
	if !ok {
		return fmt.Errorf("%s", msg)
	}
	return s.store.IncrementSignalUsage(userID, time.Now())
}

func (s *Service) CanExecuteTrade(userID int64, tradeType string) error {
	if s.IsAdmin(userID) {
		return nil
	}
	snap, err := s.Snapshot(userID)
	if err != nil {
		return fmt.Errorf("could not verify plan usage")
	}
	side := strings.ToUpper(strings.TrimSpace(tradeType))
	switch side {
	case "BUY":
		if snap.Usage.LongTrades >= snap.Limits.MaxLongTrades {
			return fmt.Errorf(
				"long trade limit reached (%d/%d this month) — upgrade your plan",
				snap.Usage.LongTrades, snap.Limits.MaxLongTrades,
			)
		}
	case "SELL":
		if snap.Usage.ShortTrades >= snap.Limits.MaxShortTrades {
			return fmt.Errorf(
				"short trade limit reached (%d/%d this month) — upgrade your plan",
				snap.Usage.ShortTrades, snap.Limits.MaxShortTrades,
			)
		}
	default:
		return fmt.Errorf("invalid trade type: %s", tradeType)
	}
	return nil
}

func (s *Service) RecordTrade(userID int64, tradeType string) error {
	if s.IsAdmin(userID) {
		return nil
	}
	if err := s.CanExecuteTrade(userID, tradeType); err != nil {
		return err
	}
	return s.store.IncrementTradeUsage(userID, tradeType, time.Now())
}
