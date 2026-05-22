package feedback

import (
	"strings"

	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
)

// Service notifies the trader and broadcasts TP/SL outcomes to signal subscribers.
type Service struct {
	inner     OutcomeNotifier
	store     *storage.PostgresStorage
	subs      *subscription.Service
	monitored map[string]bool
}

func NewService(
	inner OutcomeNotifier,
	store *storage.PostgresStorage,
	subs *subscription.Service,
	symbols []string,
) *Service {
	mon := make(map[string]bool)
	for _, s := range symbols {
		mon[strings.ToUpper(strings.TrimSpace(s))] = true
	}
	return &Service{inner: inner, store: store, subs: subs, monitored: mon}
}

func (s *Service) isMonitored(symbol string) bool {
	if len(s.monitored) == 0 {
		return true
	}
	return s.monitored[strings.ToUpper(symbol)]
}

// NotifyTradeOutcome alerts the trader; on TP/SL for monitored pairs, also alerts signal subscribers.
func (s *Service) NotifyTradeOutcome(telegramID int64, trade *models.Trade, reason string) error {
	if trade == nil {
		return nil
	}
	if s.inner != nil {
		if err := s.inner.NotifyTradeOutcome(telegramID, trade, reason); err != nil {
			logger.Warn("Trade outcome notify trader %d: %v", telegramID, err)
		}
	}
	reason = strings.ToUpper(strings.TrimSpace(reason))
	if (reason == "TP" || reason == "SL") && s.isMonitored(trade.Symbol) {
		s.broadcastToSubscribers(trade, reason, telegramID)
	}
	return nil
}

func (s *Service) broadcastToSubscribers(trade *models.Trade, reason string, excludeTelegramID int64) {
	if s.store == nil || s.inner == nil {
		return
	}
	ids, err := s.store.ListSignalSubscriberTelegramIDsForSymbol(trade.Symbol)
	if err != nil {
		logger.Error("List signal subscribers for %s: %v", trade.Symbol, err)
		return
	}
	cn, hasCommunity := s.inner.(communityNotifier)
	for _, id := range ids {
		if id == excludeTelegramID {
			continue
		}
		if s.subs != nil {
			ok, _ := s.subs.CanTrade(id)
			if !ok {
				continue
			}
		}
		var err error
		if hasCommunity {
			err = cn.NotifyCommunityOutcome(id, trade, reason)
		} else {
			err = s.inner.NotifyTradeOutcome(id, trade, reason)
		}
		if err != nil {
			logger.Warn("Community outcome notify %d: %v", id, err)
		}
	}
}

type communityNotifier interface {
	NotifyCommunityOutcome(telegramID int64, trade *models.Trade, reason string) error
}

// FormatCommunityOutcomeMessage is sent to users who received signal broadcasts.
func FormatCommunityOutcomeMessage(trade *models.Trade, reason string) string {
	if trade == nil {
		return ""
	}
	body := FormatOutcomeMessage(trade, reason)
	if body == "" {
		return ""
	}
	return "📡 *Signal result* (broadcast setup)\n\n" + body
}
