package positions

import (
	"forex-bot/internal/broker"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
)

// ListOpenForUser returns positions logged for this user (open trades only).
// When broker is set, live P/L and prices are merged for matching position ids.
func ListOpenForUser(store *storage.PostgresStorage, userID int64, b broker.Broker) ([]*models.Position, error) {
	dbPos, err := store.GetOpenPositionsByUser(userID)
	if err != nil {
		return nil, err
	}
	if len(dbPos) == 0 {
		return []*models.Position{}, nil
	}
	if b == nil {
		return dbPos, nil
	}
	brokerPos, err := b.GetOpenPositions()
	if err != nil || len(brokerPos) == 0 {
		return dbPos, nil
	}
	live := indexByPositionID(brokerPos)
	for _, p := range dbPos {
		if liveP := lookupLive(live, p); liveP != nil {
			if liveP.CurrentPrice > 0 {
				p.CurrentPrice = liveP.CurrentPrice
			}
			p.Profit = liveP.Profit
			p.ProfitPct = liveP.ProfitPct
		}
	}
	return dbPos, nil
}

func indexByPositionID(list []*models.Position) map[string]*models.Position {
	m := make(map[string]*models.Position, len(list)*2)
	for _, p := range list {
		if p.ID != "" {
			m[p.ID] = p
		}
		if p.BrokerID != "" {
			m[p.BrokerID] = p
		}
	}
	return m
}

func lookupLive(live map[string]*models.Position, p *models.Position) *models.Position {
	if p == nil {
		return nil
	}
	if x, ok := live[p.ID]; ok {
		return x
	}
	if p.BrokerID != "" {
		if x, ok := live[p.BrokerID]; ok {
			return x
		}
	}
	return nil
}
