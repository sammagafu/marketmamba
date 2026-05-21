package broker

import (
	"fmt"
	"sync"

	"forex-bot/internal/models"
)

// Broker defines the interface for broker operations
type Broker interface {
	GetBalance() (float64, error)
	GetEquity() (float64, error)
	GetOpenPositions() ([]*models.Position, error)
	OpenMarketOrder(symbol, orderType string, quantity, stopLoss, takeProfit float64) (*models.Position, error)
	ClosePosition(positionID string) error
	CloseAllPositions() error
	ModifyStopLoss(positionID string, newStopLoss float64) error
	ModifyTakeProfit(positionID string, newTakeProfit float64) error
	GetPositionByID(positionID string) (*models.Position, error)
}

// MockBroker is a mock implementation for testing
type MockBroker struct {
	balance   float64
	equity    float64
	positions map[string]*models.Position
	mu        sync.RWMutex
}

func NewMockBroker(initialBalance float64) *MockBroker {
	return &MockBroker{
		balance:   initialBalance,
		equity:    initialBalance,
		positions: make(map[string]*models.Position),
	}
}

func (m *MockBroker) GetBalance() (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.balance, nil
}

func (m *MockBroker) GetEquity() (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.equity, nil
}

func (m *MockBroker) GetOpenPositions() ([]*models.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var positions []*models.Position
	for _, pos := range m.positions {
		positions = append(positions, pos)
	}
	return positions, nil
}

func (m *MockBroker) OpenMarketOrder(symbol, orderType string, quantity, stopLoss, takeProfit float64) (*models.Position, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate inputs
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if stopLoss <= 0 || takeProfit <= 0 {
		return nil, fmt.Errorf("stop loss and take profit must be positive")
	}

	// Mock current price (for demo, use stop loss and take profit to infer entry)
	var entryPrice float64
	if orderType == "BUY" {
		entryPrice = stopLoss + (takeProfit-stopLoss)*0.5 // midpoint
	} else {
		entryPrice = takeProfit + (stopLoss-takeProfit)*0.5
	}

	positionID := fmt.Sprintf("mock_pos_%d", len(m.positions)+1)
	position := &models.Position{
		ID:         positionID,
		BrokerID:   positionID,
		Symbol:     symbol,
		Type:       orderType,
		Quantity:   quantity,
		EntryPrice: entryPrice,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
		Profit:     0,
		ProfitPct:  0,
	}

	m.positions[positionID] = position
	return position, nil
}

func (m *MockBroker) ClosePosition(positionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.positions[positionID]; !exists {
		return fmt.Errorf("position not found: %s", positionID)
	}

	delete(m.positions, positionID)
	return nil
}

func (m *MockBroker) CloseAllPositions() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.positions = make(map[string]*models.Position)
	return nil
}

func (m *MockBroker) ModifyStopLoss(positionID string, newStopLoss float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pos, exists := m.positions[positionID]
	if !exists {
		return fmt.Errorf("position not found: %s", positionID)
	}

	if newStopLoss <= 0 {
		return fmt.Errorf("stop loss must be positive")
	}

	pos.StopLoss = newStopLoss
	return nil
}

func (m *MockBroker) ModifyTakeProfit(positionID string, newTakeProfit float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pos, exists := m.positions[positionID]
	if !exists {
		return fmt.Errorf("position not found: %s", positionID)
	}

	if newTakeProfit <= 0 {
		return fmt.Errorf("take profit must be positive")
	}

	pos.TakeProfit = newTakeProfit
	return nil
}

func (m *MockBroker) GetPositionByID(positionID string) (*models.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pos, exists := m.positions[positionID]
	if !exists {
		return nil, fmt.Errorf("position not found: %s", positionID)
	}

	return pos, nil
}

// SimulatePrice updates position prices for mock trading
func (m *MockBroker) SimulatePrice(positionID string, currentPrice float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pos, exists := m.positions[positionID]
	if !exists {
		return fmt.Errorf("position not found: %s", positionID)
	}

	pos.CurrentPrice = currentPrice

	// Calculate profit
	if pos.Type == "BUY" {
		pos.Profit = (currentPrice - pos.EntryPrice) * pos.Quantity
		pos.ProfitPct = (currentPrice - pos.EntryPrice) / pos.EntryPrice * 100
	} else {
		pos.Profit = (pos.EntryPrice - currentPrice) * pos.Quantity
		pos.ProfitPct = (pos.EntryPrice - currentPrice) / pos.EntryPrice * 100
	}

	return nil
}
