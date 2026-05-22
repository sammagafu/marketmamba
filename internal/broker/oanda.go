package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forex-bot/internal/models"
)

const (
	oandaPracticeBase = "https://api-fxpractice.oanda.com"
	oandaLiveBase     = "https://api-fxtrade.oanda.com"
)

// OANDABroker implements Broker via OANDA v20 REST API.
type OANDABroker struct {
	baseURL   string
	token     string
	accountID string
	client    *http.Client
}

func NewOANDABroker(creds Credentials) (*OANDABroker, error) {
	if creds == nil {
		return nil, fmt.Errorf("OANDA credentials required")
	}
	token := strings.TrimSpace(creds["api_token"])
	accountID := strings.TrimSpace(creds["account_id"])
	if token == "" || accountID == "" {
		return nil, fmt.Errorf("OANDA api_token and account_id are required")
	}
	base := oandaLiveBase
	if practice, _ := strconv.ParseBool(creds["practice"]); practice {
		base = oandaPracticeBase
	}
	return &OANDABroker{
		baseURL:   strings.TrimRight(base, "/"),
		token:     token,
		accountID: accountID,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (o *OANDABroker) GetBalance() (float64, error) {
	summary, err := o.accountSummary()
	if err != nil {
		return 0, err
	}
	return summary.Balance, nil
}

func (o *OANDABroker) GetEquity() (float64, error) {
	summary, err := o.accountSummary()
	if err != nil {
		return 0, err
	}
	if summary.NAV > 0 {
		return summary.NAV, nil
	}
	return summary.Balance, nil
}

func (o *OANDABroker) GetOpenPositions() ([]*models.Position, error) {
	var resp struct {
		Positions []struct {
			Instrument string `json:"instrument"`
			Long       struct {
				Units        string  `json:"units"`
				AveragePrice string  `json:"averagePrice"`
				UnrealizedPL string  `json:"unrealizedPL"`
			} `json:"long"`
			Short struct {
				Units        string  `json:"units"`
				AveragePrice string  `json:"averagePrice"`
				UnrealizedPL string  `json:"unrealizedPL"`
			} `json:"short"`
		} `json:"positions"`
	}
	if err := o.getJSON("/v3/accounts/"+o.accountID+"/openPositions", &resp); err != nil {
		return nil, err
	}
	var out []*models.Position
	for _, p := range resp.Positions {
		if parseFloat(p.Long.Units) != 0 {
			out = append(out, oandaPosition(p.Instrument, "BUY", p.Long.Units, p.Long.AveragePrice, p.Long.UnrealizedPL))
		}
		if parseFloat(p.Short.Units) != 0 {
			out = append(out, oandaPosition(p.Instrument, "SELL", p.Short.Units, p.Short.AveragePrice, p.Short.UnrealizedPL))
		}
	}
	return out, nil
}

func oandaPosition(instrument, side, units, avgPrice, upl string) *models.Position {
	id := oandaToSymbol(instrument) + "|" + side
	qty := parseFloat(units)
	if qty < 0 {
		qty = -qty
	}
	return &models.Position{
		ID:           id,
		BrokerID:     id,
		Symbol:       oandaToSymbol(instrument),
		Type:         side,
		Quantity:     qty,
		EntryPrice:   parseFloat(avgPrice),
		CurrentPrice: parseFloat(avgPrice),
		Profit:       parseFloat(upl),
	}
}

func (o *OANDABroker) OpenMarketOrder(symbol, orderType string, quantity, stopLoss, takeProfit float64) (*models.Position, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	// OANDA units are in base currency units; map lot-style qty to minimum trade units.
	oandaUnits := oandaUnitsFromQuantity(quantity)
	if orderType == "SELL" {
		oandaUnits = -oandaUnits
	}
	body := map[string]interface{}{
		"order": map[string]interface{}{
			"type":       "MARKET",
			"instrument": symbolToOANDA(symbol),
			"units":      strconv.FormatInt(oandaUnits, 10),
			"stopLossOnFill": map[string]string{
				"price": fmt.Sprintf("%.5f", stopLoss),
			},
			"takeProfitOnFill": map[string]string{
				"price": fmt.Sprintf("%.5f", takeProfit),
			},
		},
	}
	var resp struct {
		OrderFillTransaction struct {
			ID         string `json:"id"`
			Instrument string `json:"instrument"`
			Units      string `json:"units"`
			Price      string `json:"price"`
		} `json:"orderFillTransaction"`
	}
	if err := o.postJSON("/v3/accounts/"+o.accountID+"/orders", body, &resp); err != nil {
		return nil, err
	}
	fill := resp.OrderFillTransaction
	if fill.ID == "" {
		return nil, fmt.Errorf("OANDA: order submitted but no fill transaction")
	}
	side := orderType
	qty := parseFloat(fill.Units)
	if qty < 0 {
		qty = -qty
		side = "SELL"
	}
	return &models.Position{
		ID:         fill.ID,
		BrokerID:   fill.ID,
		Symbol:     oandaToSymbol(fill.Instrument),
		Type:       side,
		Quantity:   qty,
		EntryPrice: parseFloat(fill.Price),
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
	}, nil
}

func (o *OANDABroker) ClosePosition(positionID string) error {
	parts := strings.Split(positionID, "|")
	if len(parts) != 2 {
		return fmt.Errorf("OANDA close: unrecognized position id %s (expected SYMBOL|SIDE)", positionID)
	}
	instrument := symbolToOANDA(parts[0])
	side := parts[1]
	var body map[string]string
	if side == "BUY" {
		body = map[string]string{"longUnits": "ALL"}
	} else {
		body = map[string]string{"shortUnits": "ALL"}
	}
	path := fmt.Sprintf("/v3/accounts/%s/positions/%s/close", o.accountID, instrument)
	return o.putJSON(path, body, nil)
}

func (o *OANDABroker) CloseAllPositions() error {
	positions, err := o.GetOpenPositions()
	if err != nil {
		return err
	}
	for _, p := range positions {
		if err := o.ClosePosition(p.ID); err != nil {
			return err
		}
	}
	return nil
}

func (o *OANDABroker) ModifyStopLoss(positionID string, newStopLoss float64) error {
	return fmt.Errorf("OANDA modify stop loss: use OANDA client — not implemented")
}

func (o *OANDABroker) ModifyTakeProfit(positionID string, newTakeProfit float64) error {
	return fmt.Errorf("OANDA modify take profit: use OANDA client — not implemented")
}

func (o *OANDABroker) GetPositionByID(positionID string) (*models.Position, error) {
	positions, err := o.GetOpenPositions()
	if err != nil {
		return nil, err
	}
	for _, p := range positions {
		if p.ID == positionID || p.BrokerID == positionID {
			return p, nil
		}
	}
	return nil, fmt.Errorf("position not found: %s", positionID)
}

type oandaAccountSummary struct {
	Balance float64
	NAV     float64
}

func (o *OANDABroker) accountSummary() (*oandaAccountSummary, error) {
	var resp struct {
		Account struct {
			Balance string `json:"balance"`
			NAV     string `json:"NAV"`
		} `json:"account"`
	}
	if err := o.getJSON("/v3/accounts/"+o.accountID+"/summary", &resp); err != nil {
		return nil, err
	}
	return &oandaAccountSummary{
		Balance: parseFloat(resp.Account.Balance),
		NAV:     parseFloat(resp.Account.NAV),
	}, nil
}

func (o *OANDABroker) getJSON(path string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, o.baseURL+path, nil)
	if err != nil {
		return err
	}
	return o.do(req, out)
}

func (o *OANDABroker) postJSON(path string, body interface{}, out interface{}) error {
	return o.jsonRequest(http.MethodPost, path, body, out)
}

func (o *OANDABroker) putJSON(path string, body interface{}, out interface{}) error {
	return o.jsonRequest(http.MethodPut, path, body, out)
}

func (o *OANDABroker) jsonRequest(method, path string, body interface{}, out interface{}) error {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, o.baseURL+path, buf)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return o.do(req, out)
}

func (o *OANDABroker) do(req *http.Request, out interface{}) error {
	req.Header.Set("Authorization", "Bearer "+o.token)
	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("OANDA API %s %s: %s", req.Method, req.URL.Path, truncate(string(data), 300))
	}
	if out == nil {
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func symbolToOANDA(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.Contains(s, "_") {
		return s
	}
	if len(s) == 6 {
		return s[:3] + "_" + s[3:]
	}
	if strings.Contains(s, "BTC") {
		return "BTC_USD"
	}
	return s
}

func oandaToSymbol(instrument string) string {
	return strings.ReplaceAll(instrument, "_", "")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// oandaUnitsFromQuantity maps internal lot sizing to OANDA unit count (min 1).
func oandaUnitsFromQuantity(quantity float64) int64 {
	if quantity <= 0 {
		return 1
	}
	u := int64(quantity * 1000)
	if u < 1 {
		u = 1
	}
	return u
}
