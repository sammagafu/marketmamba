package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateID creates a random ID for database records
func GenerateID(prefix string) string {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return prefix + "_error"
	}
	return prefix + "_" + hex.EncodeToString(randomBytes)
}

// FormatCurrency formats a float as currency
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// FormatPercent formats a float as percentage
func FormatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value*100)
}

// ParsePrice parses a price string to float64
func ParsePrice(s string) (float64, error) {
	var price float64
	_, err := fmt.Sscanf(s, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("invalid price format: %s", s)
	}
	return price, nil
}

// IsValidSymbol checks if symbol is valid
func IsValidSymbol(symbol string) bool {
	symbol = strings.ToUpper(symbol)
	validSymbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "NZDUSD", "USDCAD",
		"EURJPY", "EURGBP", "GBPJPY", "CHFJPY",
		"BTCUSD", "ETHUSD", "XAUUSD",
		"US500", "USTEC", "GER40", "UK100", "VOL75",
	}

	for _, valid := range validSymbols {
		if symbol == valid {
			return true
		}
	}
	return false
}

// CalculateRiskRewardRatio computes the risk-reward ratio
func CalculateRiskRewardRatio(entryPrice, stopLoss, takeProfit float64) (float64, error) {
	if entryPrice <= 0 || stopLoss <= 0 || takeProfit <= 0 {
		return 0, fmt.Errorf("all prices must be positive")
	}

	riskPips := abs(entryPrice - stopLoss)
	rewardPips := abs(takeProfit - entryPrice)

	if riskPips <= 0 {
		return 0, fmt.Errorf("invalid stop loss")
	}

	return rewardPips / riskPips, nil
}

// CalculateLotSize computes lot size based on risk
func CalculateLotSize(accountBalance, riskPercent, stopLossDistance float64) float64 {
	if accountBalance <= 0 || riskPercent <= 0 || stopLossDistance <= 0 {
		return 0
	}

	riskAmount := accountBalance * riskPercent
	return riskAmount / stopLossDistance
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
