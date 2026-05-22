package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Provider fetches live quotes for a symbol.
type Provider interface {
	FetchQuote(ctx context.Context, symbol string) (*Quote, error)
	Name() string
}

// CompositeProvider routes symbols to the best available feed.
type CompositeProvider struct {
	twelveKey string
	client    *http.Client
}

func NewCompositeProvider(twelveDataAPIKey string) *CompositeProvider {
	return &CompositeProvider{
		twelveKey: strings.TrimSpace(twelveDataAPIKey),
		client:    &http.Client{Timeout: 12 * time.Second},
	}
}

func (c *CompositeProvider) Name() string {
	if c.twelveKey != "" {
		return "twelvedata+free"
	}
	return "live-free"
}

func (c *CompositeProvider) FetchQuote(ctx context.Context, symbol string) (*Quote, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if c.twelveKey != "" {
		if q, err := c.fetchTwelveData(ctx, sym); err == nil {
			return q, nil
		}
	}
	switch {
	case strings.Contains(sym, "BTC"):
		return c.fetchCoinGecko(ctx, sym)
	default:
		return c.fetchFrankfurter(ctx, sym)
	}
}

func (c *CompositeProvider) fetchFrankfurter(ctx context.Context, symbol string) (*Quote, error) {
	// EURUSD = USD per 1 EUR from Frankfurter (base EUR).
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.frankfurter.app/latest?from=EUR&to=USD", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("frankfurter status %d", resp.StatusCode)
	}
	var payload struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	mid, ok := payload.Rates["USD"]
	if !ok || mid <= 0 {
		return nil, fmt.Errorf("frankfurter: missing USD rate")
	}
	spread := pipSpread(symbol, mid)
	return &Quote{
		Symbol:    normalizeForexSymbol(symbol),
		Bid:       mid - spread/2,
		Ask:       mid + spread/2,
		Mid:       mid,
		Source:    "frankfurter",
		FetchedAt: time.Now().UTC(),
	}, nil
}

func (c *CompositeProvider) fetchCoinGecko(ctx context.Context, symbol string) (*Quote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko status %d", resp.StatusCode)
	}
	var payload map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	mid := payload["bitcoin"]["usd"]
	if mid <= 0 {
		return nil, fmt.Errorf("coingecko: invalid BTC price")
	}
	spread := mid * 0.00015
	return &Quote{
		Symbol:    "BTCUSD",
		Bid:       mid - spread/2,
		Ask:       mid + spread/2,
		Mid:       mid,
		Source:    "coingecko",
		FetchedAt: time.Now().UTC(),
	}, nil
}

func (c *CompositeProvider) fetchTwelveData(ctx context.Context, symbol string) (*Quote, error) {
	tdSym := toTwelveSymbol(symbol)
	url := fmt.Sprintf("https://api.twelvedata.com/quote?symbol=%s&apikey=%s", tdSym, c.twelveKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twelvedata status %d", resp.StatusCode)
	}
	var payload struct {
		Symbol string `json:"symbol"`
		Bid    string `json:"bid"`
		Ask    string `json:"ask"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Status == "error" {
		return nil, fmt.Errorf("twelvedata error")
	}
	bid := parseFloat(payload.Bid)
	ask := parseFloat(payload.Ask)
	if bid <= 0 || ask <= 0 {
		return nil, fmt.Errorf("twelvedata: invalid quote")
	}
	return &Quote{
		Symbol:    strings.ToUpper(symbol),
		Bid:       bid,
		Ask:       ask,
		Mid:       (bid + ask) / 2,
		Source:    "twelvedata",
		FetchedAt: time.Now().UTC(),
	}, nil
}

// SeedBars loads historical closes when Twelve Data key is configured.
func (c *CompositeProvider) SeedBars(ctx context.Context, symbol string, count int) ([]float64, error) {
	if c.twelveKey == "" || count <= 0 {
		return nil, fmt.Errorf("seeding unavailable")
	}
	tdSym := toTwelveSymbol(symbol)
	url := fmt.Sprintf(
		"https://api.twelvedata.com/time_series?symbol=%s&interval=1min&outputsize=%d&apikey=%s",
		tdSym, count, c.twelveKey,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload struct {
		Values []struct {
			Close string `json:"close"`
		} `json:"values"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Status == "error" || len(payload.Values) == 0 {
		return nil, fmt.Errorf("twelvedata time_series failed")
	}
	// API returns newest first.
	closes := make([]float64, 0, len(payload.Values))
	for i := len(payload.Values) - 1; i >= 0; i-- {
		c := parseFloat(payload.Values[i].Close)
		if c > 0 {
			closes = append(closes, c)
		}
	}
	return closes, nil
}

func pipSpread(symbol string, mid float64) float64 {
	if strings.Contains(strings.ToUpper(symbol), "BTC") {
		return mid * 0.00015
	}
	// ~1 pip on EURUSD.
	return 0.0001
}

func normalizeForexSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if s == "" {
		return "EURUSD"
	}
	return s
}

func toTwelveSymbol(symbol string) string {
	s := strings.ToUpper(symbol)
	if strings.Contains(s, "BTC") {
		return "BTC/USD"
	}
	if len(s) == 6 {
		return s[:3] + "/" + s[3:]
	}
	return s
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f
}
