package trading

import (
	"fmt"
	"time"

	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/utils"
)

// TradeLog persists opens/closes to Postgres and writes structured logs.
type TradeLog struct {
	store *storage.PostgresStorage
}

func NewTradeLog(store *storage.PostgresStorage) *TradeLog {
	return &TradeLog{store: store}
}

// RecordOpen saves trade + position rows (position id = broker position id).
func (tl *TradeLog) RecordOpen(userID int64, brokerPos *models.Position, source string) (*models.Trade, error) {
	if tl == nil || tl.store == nil || brokerPos == nil {
		return nil, fmt.Errorf("invalid trade log input")
	}
	now := time.Now()
	riskAmt := abs((brokerPos.EntryPrice - brokerPos.StopLoss) * brokerPos.Quantity)
	rewardAmt := abs((brokerPos.TakeProfit - brokerPos.EntryPrice) * brokerPos.Quantity)
	rr := 0.0
	if riskAmt > 0 {
		rr = rewardAmt / riskAmt
	}
	trade := &models.Trade{
		ID:              utils.GenerateID("trade"),
		UserID:          userID,
		Symbol:          brokerPos.Symbol,
		Type:            brokerPos.Type,
		EntryPrice:      brokerPos.EntryPrice,
		Quantity:        brokerPos.Quantity,
		StopLoss:        brokerPos.StopLoss,
		TakeProfit:      brokerPos.TakeProfit,
		RiskAmount:      riskAmt,
		RewardAmount:    rewardAmt,
		RiskRewardRatio: rr,
		Status:          "OPEN",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := tl.store.CreateTrade(trade); err != nil {
		return nil, err
	}
	pos := &models.Position{
		ID:         brokerPos.ID,
		TradeID:    trade.ID,
		BrokerID:   brokerPos.BrokerID,
		UserID:     userID,
		Symbol:     brokerPos.Symbol,
		Type:       brokerPos.Type,
		Quantity:   brokerPos.Quantity,
		EntryPrice: brokerPos.EntryPrice,
		StopLoss:   brokerPos.StopLoss,
		TakeProfit: brokerPos.TakeProfit,
		UpdatedAt:  now,
	}
	if pos.BrokerID == "" {
		pos.BrokerID = brokerPos.ID
	}
	if err := tl.store.CreatePosition(pos); err != nil {
		logger.Error("Failed to save position %s for trade %s: %v", pos.ID, trade.ID, err)
	}
	logger.Info(
		"TRADE OPEN user=%d trade=%s %s %s qty=%.2f entry=%.5f SL=%.5f TP=%.5f source=%s pos=%s",
		userID, trade.ID, trade.Symbol, trade.Type, trade.Quantity, trade.EntryPrice,
		trade.StopLoss, trade.TakeProfit, source, pos.ID,
	)
	return trade, nil
}

// RecordClose closes a trade by broker position id.
func (tl *TradeLog) RecordClose(userID int64, brokerPositionID string, exitPrice float64, reason string) (*models.Trade, error) {
	if tl == nil || tl.store == nil {
		return nil, fmt.Errorf("invalid trade log input")
	}
	trade, err := tl.store.GetTradeByBrokerPositionID(userID, brokerPositionID)
	if err != nil {
		return nil, err
	}
	if trade == nil {
		return nil, fmt.Errorf("no open trade for position %s", brokerPositionID)
	}
	now := time.Now()
	var profit float64
	if trade.Type == "BUY" {
		profit = (exitPrice - trade.EntryPrice) * trade.Quantity
	} else {
		profit = (trade.EntryPrice - exitPrice) * trade.Quantity
	}
	trade.Status = "CLOSED"
	trade.ExitPrice = &exitPrice
	trade.ExitTime = &now
	trade.Profit = &profit
	trade.ClosureReason = &reason
	trade.UpdatedAt = now
	if err := tl.store.UpdateTrade(trade); err != nil {
		return nil, err
	}
	_ = tl.store.DeletePosition(brokerPositionID)
	logger.Info(
		"TRADE CLOSE user=%d trade=%s %s %s exit=%.5f profit=%.2f reason=%s pos=%s",
		userID, trade.ID, trade.Symbol, trade.Type, exitPrice, profit, reason, brokerPositionID,
	)
	return trade, nil
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
