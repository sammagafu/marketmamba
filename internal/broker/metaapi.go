package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"forex-bot/internal/models"
)

const (
	metaAPIProvisionBase = "https://mt-provisioning-api-v1.%s.agiliumtrade.ai"
	metaAPIClientBase    = "https://mt-client-api-v1.%s.agiliumtrade.ai"
	metaAPIDefaultRegion = "new-york"
)

// MetaAPIBroker executes trades via MetaAPI cloud (MT4/MT5). Users supply their own
// broker server/login/password (e.g. Deriv-Demo, Exness-Demo, ICMarketsSC-Demo).
type MetaAPIBroker struct {
	token  string
	region string
	creds  Credentials
	client *http.Client

	provisionBase string // test override
	clientBase    string // test override

	mu             sync.Mutex
	ready          bool
	accountID      string
	waitDeployedFn func(accountID string) error // tests only
}

func NewMetaAPIBroker(creds Credentials) (*MetaAPIBroker, error) {
	if err := ValidateMetaAPICredentials(creds); err != nil {
		return nil, err
	}
	return &MetaAPIBroker{
		token:  metaAPIToken(creds),
		region: metaAPIRegion(creds),
		creds:  cloneCredentials(creds),
		client: &http.Client{Timeout: 45 * time.Second},
	}, nil
}

// ValidateMetaAPICredentials ensures MetaAPI token plus either cloud account id or MT login details.
func ValidateMetaAPICredentials(creds Credentials) error {
	if creds == nil {
		return fmt.Errorf("MetaAPI credentials required")
	}
	if metaAPIToken(creds) == "" {
		return fmt.Errorf("MetaAPI token is required (from app.metaapi.cloud)")
	}
	if metaAPICloudAccountID(creds) != "" {
		return nil
	}
	if strings.TrimSpace(creds["login"]) == "" {
		return fmt.Errorf("MT login is required (or MetaAPI account id)")
	}
	if strings.TrimSpace(creds["password"]) == "" {
		return fmt.Errorf("MT password is required (or MetaAPI account id)")
	}
	if strings.TrimSpace(creds["server"]) == "" {
		return fmt.Errorf("MT server is required (e.g. Deriv-Demo)")
	}
	return nil
}

func (m *MetaAPIBroker) GetBalance() (float64, error) {
	info, err := m.accountInformation()
	if err != nil {
		return 0, err
	}
	return info.Balance, nil
}

func (m *MetaAPIBroker) GetEquity() (float64, error) {
	info, err := m.accountInformation()
	if err != nil {
		return 0, err
	}
	if info.Equity > 0 {
		return info.Equity, nil
	}
	return info.Balance, nil
}

func (m *MetaAPIBroker) GetOpenPositions() ([]*models.Position, error) {
	if err := m.ensureReady(); err != nil {
		return nil, err
	}
	var raw []metaAPIPosition
	if err := m.clientGET(fmt.Sprintf("/users/current/accounts/%s/positions", m.accountID), &raw); err != nil {
		return nil, err
	}
	out := make([]*models.Position, 0, len(raw))
	for _, p := range raw {
		if p.Volume == 0 {
			continue
		}
		side := "BUY"
		if strings.EqualFold(p.Type, "POSITION_TYPE_SELL") || strings.Contains(strings.ToUpper(p.Type), "SELL") {
			side = "SELL"
		}
		id := p.ID
		if id == "" {
			id = fmt.Sprintf("%s|%s", metaAPIToSymbol(p.Symbol), side)
		}
		out = append(out, &models.Position{
			ID:           id,
			BrokerID:     id,
			Symbol:       metaAPIToSymbol(p.Symbol),
			Type:         side,
			Quantity:     p.Volume,
			EntryPrice:   p.OpenPrice,
			CurrentPrice: p.CurrentPrice,
			StopLoss:     p.StopLoss,
			TakeProfit:   p.TakeProfit,
			Profit:       p.Profit,
		})
	}
	return out, nil
}

func (m *MetaAPIBroker) OpenMarketOrder(symbol, orderType string, quantity, stopLoss, takeProfit float64) (*models.Position, error) {
	if err := m.ensureReady(); err != nil {
		return nil, err
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	action := "ORDER_TYPE_BUY"
	if strings.EqualFold(orderType, "SELL") {
		action = "ORDER_TYPE_SELL"
	}
	vol := quantity
	if vol < 0.01 {
		vol = 0.01
	}
	var lastErr error
	for _, sym := range metaAPISymbolCandidates(symbol) {
		body := map[string]interface{}{
			"actionType": action,
			"symbol":     sym,
			"volume":     vol,
		}
		if stopLoss > 0 {
			body["stopLoss"] = stopLoss
		}
		if takeProfit > 0 {
			body["takeProfit"] = takeProfit
		}
		var resp metaAPITradeResponse
		err := m.clientPOST(fmt.Sprintf("/users/current/accounts/%s/trade", m.accountID), body, &resp)
		if err != nil {
			lastErr = err
			continue
		}
		if !resp.isSuccess() {
			lastErr = fmt.Errorf("MetaAPI trade: %s", resp.Message)
			continue
		}
		positions, err := m.GetOpenPositions()
		if err == nil {
			for _, p := range positions {
				if strings.EqualFold(p.Symbol, metaAPIToSymbol(sym)) || strings.EqualFold(p.Symbol, symbol) {
					return p, nil
				}
			}
		}
		return &models.Position{
			ID:       resp.PositionID,
			BrokerID: resp.PositionID,
			Symbol:   metaAPIToSymbol(sym),
			Type:     strings.ToUpper(orderType),
			Quantity: vol,
		}, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("could not open order for symbol %s", symbol)
}

func (m *MetaAPIBroker) ClosePosition(positionID string) error {
	if err := m.ensureReady(); err != nil {
		return err
	}
	body := map[string]interface{}{
		"actionType": "POSITION_CLOSE_ID",
		"positionId": positionID,
	}
	var resp metaAPITradeResponse
	if err := m.clientPOST(fmt.Sprintf("/users/current/accounts/%s/trade", m.accountID), body, &resp); err != nil {
		return err
	}
	if !resp.isSuccess() {
		return fmt.Errorf("MetaAPI close: %s", resp.Message)
	}
	return nil
}

func (m *MetaAPIBroker) CloseAllPositions() error {
	positions, err := m.GetOpenPositions()
	if err != nil {
		return err
	}
	for _, p := range positions {
		if err := m.ClosePosition(p.ID); err != nil {
			return err
		}
	}
	return nil
}

func (m *MetaAPIBroker) ModifyStopLoss(positionID string, newStopLoss float64) error {
	return m.modifyPosition(positionID, newStopLoss, 0)
}

func (m *MetaAPIBroker) ModifyTakeProfit(positionID string, newTakeProfit float64) error {
	return m.modifyPosition(positionID, 0, newTakeProfit)
}

func (m *MetaAPIBroker) modifyPosition(positionID string, sl, tp float64) error {
	if err := m.ensureReady(); err != nil {
		return err
	}
	body := map[string]interface{}{
		"actionType": "POSITION_MODIFY",
		"positionId": positionID,
	}
	if sl > 0 {
		body["stopLoss"] = sl
	}
	if tp > 0 {
		body["takeProfit"] = tp
	}
	var resp metaAPITradeResponse
	if err := m.clientPOST(fmt.Sprintf("/users/current/accounts/%s/trade", m.accountID), body, &resp); err != nil {
		return err
	}
	if !resp.isSuccess() {
		return fmt.Errorf("MetaAPI modify: %s", resp.Message)
	}
	return nil
}

func (m *MetaAPIBroker) GetPositionByID(positionID string) (*models.Position, error) {
	positions, err := m.GetOpenPositions()
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

func (m *MetaAPIBroker) ensureReady() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ready && m.accountID != "" {
		return nil
	}
	id := metaAPICloudAccountID(m.creds)
	if id == "" {
		var err error
		id, err = m.findOrCreateAccount()
		if err != nil {
			return err
		}
	}
	if err := m.waitDeployed(id); err != nil {
		return err
	}
	m.accountID = id
	m.ready = true
	return nil
}

func (m *MetaAPIBroker) findOrCreateAccount() (string, error) {
	login := strings.TrimSpace(m.creds["login"])
	server := strings.TrimSpace(m.creds["server"])
	accounts, err := m.listAccounts()
	if err != nil {
		return "", err
	}
	for _, a := range accounts {
		if strings.TrimSpace(a.Login) == login && strings.EqualFold(strings.TrimSpace(a.Server), server) {
			return a.accountID(), nil
		}
	}
	return m.createAccount()
}

func (m *MetaAPIBroker) createAccount() (string, error) {
	login := strings.TrimSpace(m.creds["login"])
	platform := strings.ToLower(strings.TrimSpace(m.creds["platform"]))
	if platform == "" {
		platform = "mt5"
	}
	name := strings.TrimSpace(m.creds["name"])
	if name == "" {
		name = fmt.Sprintf("marketmamba-%s", login)
	}
	body := map[string]interface{}{
		"login":    login,
		"password": strings.TrimSpace(m.creds["password"]),
		"server":   strings.TrimSpace(m.creds["server"]),
		"name":     name,
		"platform": platform,
		"magic":    202602,
	}
	if k := strings.TrimSpace(m.creds["keywords"]); k != "" {
		parts := strings.Split(k, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		body["keywords"] = parts
	}
	var created metaAPIAccount
	if err := m.provisionPOST("/users/current/accounts", body, &created); err != nil {
		return "", err
	}
	id := created.accountID()
	if id == "" {
		return "", fmt.Errorf("MetaAPI did not return account id")
	}
	return id, nil
}

func (m *MetaAPIBroker) listAccounts() ([]metaAPIAccount, error) {
	var accounts []metaAPIAccount
	if err := m.provisionGET("/users/current/accounts", &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (m *MetaAPIBroker) waitDeployed(accountID string) error {
	if m.waitDeployedFn != nil {
		return m.waitDeployedFn(accountID)
	}
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		var acc metaAPIAccount
		if err := m.provisionGET("/users/current/accounts/"+accountID, &acc); err != nil {
			return err
		}
		state := strings.ToUpper(strings.TrimSpace(acc.State))
		switch state {
		case "DEPLOYED":
			return nil
		case "DEPLOY_FAILED", "REMOVED":
			return fmt.Errorf("MetaAPI account state %s", state)
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("MetaAPI account %s not deployed in time — check app.metaapi.cloud", accountID)
}

type metaAPIAccountInfo struct {
	Balance float64 `json:"balance"`
	Equity  float64 `json:"equity"`
}

func (m *MetaAPIBroker) accountInformation() (*metaAPIAccountInfo, error) {
	if err := m.ensureReady(); err != nil {
		return nil, err
	}
	var info metaAPIAccountInfo
	if err := m.clientGET(fmt.Sprintf("/users/current/accounts/%s/account-information", m.accountID), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

type metaAPIAccount struct {
	ID     string `json:"_id"`
	AltID  string `json:"id"`
	Login  string `json:"login"`
	Server string `json:"server"`
	State  string `json:"state"`
}

func (a metaAPIAccount) accountID() string {
	if a.ID != "" {
		return a.ID
	}
	return a.AltID
}

type metaAPIPosition struct {
	ID           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Type         string  `json:"type"`
	Volume       float64 `json:"volume"`
	OpenPrice    float64 `json:"openPrice"`
	CurrentPrice float64 `json:"currentPrice"`
	StopLoss     float64 `json:"stopLoss"`
	TakeProfit   float64 `json:"takeProfit"`
	Profit       float64 `json:"profit"`
}

type metaAPITradeResponse struct {
	NumericCode int    `json:"numericCode"`
	StringCode  string `json:"stringCode"`
	Message     string `json:"message"`
	PositionID  string `json:"positionId"`
	OrderID     string `json:"orderId"`
}

func (r metaAPITradeResponse) isSuccess() bool {
	if r.StringCode == "TRADE_RETCODE_DONE" {
		return true
	}
	return r.NumericCode == 10009
}

func (m *MetaAPIBroker) provisionBaseURL() string {
	if m.provisionBase != "" {
		return strings.TrimRight(m.provisionBase, "/")
	}
	return fmt.Sprintf(metaAPIProvisionBase, m.region)
}

func (m *MetaAPIBroker) clientBaseURL() string {
	if m.clientBase != "" {
		return strings.TrimRight(m.clientBase, "/")
	}
	return fmt.Sprintf(metaAPIClientBase, m.region)
}

func (m *MetaAPIBroker) provisionGET(path string, out interface{}) error {
	return m.apiGET(m.provisionBaseURL()+path, out)
}

func (m *MetaAPIBroker) provisionPOST(path string, body interface{}, out interface{}) error {
	return m.apiPOST(m.provisionBaseURL()+path, body, out)
}

func (m *MetaAPIBroker) clientGET(path string, out interface{}) error {
	return m.apiGET(m.clientBaseURL()+path, out)
}

func (m *MetaAPIBroker) clientPOST(path string, body interface{}, out interface{}) error {
	return m.apiPOST(m.clientBaseURL()+path, body, out)
}

func (m *MetaAPIBroker) apiGET(url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	return m.do(req, out)
}

func (m *MetaAPIBroker) apiPOST(url string, body interface{}, out interface{}) error {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return m.do(req, out)
}

func (m *MetaAPIBroker) do(req *http.Request, out interface{}) error {
	req.Header.Set("auth-token", m.token)
	req.Header.Set("Accept", "application/json")
	if req.Method == http.MethodPost {
		req.Header.Set("transaction-id", fmt.Sprintf("mm-%d", time.Now().UnixNano()))
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("MetaAPI %s %s: %s", req.Method, req.URL.Path, truncate(string(data), 400))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func metaAPIToken(creds Credentials) string {
	for _, k := range []string{"metaapi_token", "token", "api_token"} {
		if v := strings.TrimSpace(creds[k]); v != "" {
			return v
		}
	}
	return ""
}

func metaAPIRegion(creds Credentials) string {
	if r := strings.TrimSpace(creds["region"]); r != "" {
		return r
	}
	return metaAPIDefaultRegion
}

func metaAPICloudAccountID(creds Credentials) string {
	for _, k := range []string{"metaapi_account_id", "account_id"} {
		v := strings.TrimSpace(creds[k])
		if isMetaAPICloudID(v) {
			return v
		}
	}
	return ""
}

func isMetaAPICloudID(s string) bool {
	return len(s) >= 32 && strings.Contains(s, "-")
}

func metaAPISymbolCandidates(symbol string) []string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	switch s {
	case "EURUSD":
		return []string{"frxEURUSD", "EURUSD", "EURUSDm"}
	case "GBPUSD":
		return []string{"frxGBPUSD", "GBPUSD"}
	case "USDJPY":
		return []string{"frxUSDJPY", "USDJPY"}
	case "BTCUSD":
		return []string{"cryBTCUSD", "BTCUSD"}
	default:
		if strings.HasPrefix(s, "FRX") || strings.HasPrefix(s, "CRY") {
			return []string{s}
		}
		return []string{"frx" + s, s}
	}
}

func metaAPIToSymbol(sym string) string {
	s := strings.TrimSpace(sym)
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "frx") && len(s) > 3 {
		return strings.ToUpper(s[3:])
	}
	if strings.HasPrefix(low, "cry") && len(s) > 3 {
		return strings.ToUpper(s[3:])
	}
	return strings.ToUpper(s)
}

func cloneCredentials(c Credentials) Credentials {
	out := Credentials{}
	for k, v := range c {
		out[k] = v
	}
	return out
}

func credsHintServer(creds Credentials) string {
	if creds == nil {
		return ""
	}
	return strings.TrimSpace(creds["server"])
}
