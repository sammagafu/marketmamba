package risk

import (
	"testing"
	"time"

	"forex-bot/internal/models"
)

func TestCalculateLotSize(t *testing.T) {
	settings := &models.RiskSettings{
		MaxRiskPerTrade: 0.005, // 0.5%
	}
	validator := NewRiskValidator(settings)

	balance := 10000.0
	entryPrice := 1.1000
	stopLoss := 1.0950

	lotSize, err := validator.CalculateLotSize(balance, entryPrice, stopLoss)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lotSize <= 0 {
		t.Errorf("lot size should be positive, got %f", lotSize)
	}

	expectedRisk := balance * settings.MaxRiskPerTrade
	pipsRisk := entryPrice - stopLoss
	expectedLotSize := expectedRisk / pipsRisk

	if lotSize != expectedLotSize {
		t.Errorf("lot size mismatch: got %f, expected %f", lotSize, expectedLotSize)
	}
}

func TestValidateTradeSignal_BuyOrder(t *testing.T) {
	settings := &models.RiskSettings{
		MaxRiskPerTrade:  0.005,
		MaxDailyLoss:     0.02,
		MaxOpenTrades:    2,
		MaxTradesPerDay:  10,
		RiskRewardRatio:  1.0,
	}
	validator := NewRiskValidator(settings)

	signal := &models.TradeSignal{
		Symbol:          "EURUSD",
		Type:            "BUY",
		Strength:        0.8,
		StopLoss:        1.0950,
		TakeProfit:      1.1050,
		RiskRewardRatio: 1.0,
		TriggeredAt:     time.Now(),
	}

	err := validator.ValidateTradeSignal(signal, 10000, 1, 5, 0, false)
	if err != nil {
		t.Errorf("unexpected error for valid signal: %v", err)
	}
}

func TestValidateTradeSignal_Paused(t *testing.T) {
	settings := &models.RiskSettings{}
	validator := NewRiskValidator(settings)

	signal := &models.TradeSignal{}

	err := validator.ValidateTradeSignal(signal, 10000, 0, 0, 0, true)
	if err == nil {
		t.Errorf("expected error when trading is paused")
	}
}

func TestValidateTradeSignal_MaxOpenTrades(t *testing.T) {
	settings := &models.RiskSettings{
		MaxOpenTrades: 2,
	}
	validator := NewRiskValidator(settings)

	signal := &models.TradeSignal{
		Type:       "BUY",
		StopLoss:   1.0950,
		TakeProfit: 1.1050,
		Strength:   0.8,
	}

	err := validator.ValidateTradeSignal(signal, 10000, 2, 0, 0, false)
	if err == nil {
		t.Errorf("expected error when max open trades reached")
	}
}

func TestValidateTradeSignal_InvalidOrderType(t *testing.T) {
	settings := &models.RiskSettings{}
	validator := NewRiskValidator(settings)

	signal := &models.TradeSignal{
		Type:       "INVALID",
		StopLoss:   1.0950,
		TakeProfit: 1.1050,
	}

	err := validator.ValidateTradeSignal(signal, 10000, 0, 0, 0, false)
	if err == nil {
		t.Errorf("expected error for invalid order type")
	}
}

func TestCanOpenMoreTrades(t *testing.T) {
	settings := &models.RiskSettings{
		MaxOpenTrades:  2,
		MaxTradesPerDay: 10,
	}
	validator := NewRiskValidator(settings)

	tests := []struct {
		openCount   int
		todayCount  int
		expected    bool
		description string
	}{
		{0, 0, true, "no limits hit"},
		{1, 5, true, "both under limits"},
		{2, 10, false, "both at limits"},
		{3, 5, false, "open trades exceeded"},
		{1, 10, false, "daily trades exceeded"},
	}

	for _, tt := range tests {
		result := validator.CanOpenMoreTrades(tt.openCount, tt.todayCount)
		if result != tt.expected {
			t.Errorf("%s: got %v, expected %v", tt.description, result, tt.expected)
		}
	}
}
