package payments

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"crypto/rand"
	"time"
)

const binancePayBase = "https://bpay.binanceapi.com"

type BinancePayClient struct {
	APIKey     string
	SecretKey  string
	CertSN     string
	HTTPClient *http.Client
}

type createOrderRequest struct {
	Env struct {
		TerminalType string `json:"terminalType"`
	} `json:"env"`
	MerchantTradeNo string  `json:"merchantTradeNo"`
	OrderAmount     float64 `json:"orderAmount"`
	Currency        string  `json:"currency"`
	Goods           struct {
		GoodsType        string `json:"goodsType"`
		GoodsCategory    string `json:"goodsCategory"`
		ReferenceGoodsID string `json:"referenceGoodsId"`
		GoodsName        string `json:"goodsName"`
		GoodsDetail      string `json:"goodsDetail"`
	} `json:"goods"`
}

type createOrderResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Data   struct {
		PrepayID      string `json:"prepayId"`
		UniversalURL  string `json:"universalUrl"`
		CheckoutURL   string `json:"checkoutUrl"`
		QRCodeLink    string `json:"qrcodeLink"`
		Deeplink      string `json:"deeplink"`
	} `json:"data"`
	ErrorMessage string `json:"errorMessage"`
}

func (c *BinancePayClient) Enabled() bool {
	return c != nil && c.APIKey != "" && c.SecretKey != "" && c.CertSN != ""
}

func (c *BinancePayClient) CreateUSDTOrder(merchantTradeNo string, amount float64, goodsName string) (prepayID, checkoutURL string, err error) {
	if !c.Enabled() {
		return "", "", fmt.Errorf("binance pay not configured")
	}
	var req createOrderRequest
	req.Env.TerminalType = "WEB"
	req.MerchantTradeNo = merchantTradeNo
	req.OrderAmount = amount
	req.Currency = "USDT"
	req.Goods.GoodsType = "01"
	req.Goods.GoodsCategory = "Z000"
	req.Goods.ReferenceGoodsID = "mm-sub"
	req.Goods.GoodsName = goodsName
	req.Goods.GoodsDetail = goodsName

	body, _ := json.Marshal(req)
	var resp createOrderResponse
	if err := c.post("/binancepay/openapi/v3/order", body, &resp); err != nil {
		return "", "", err
	}
	if resp.Status != "SUCCESS" && resp.Code != "000000" {
		msg := resp.ErrorMessage
		if msg == "" {
			msg = resp.Code
		}
		return "", "", fmt.Errorf("binance pay: %s", msg)
	}
	url := resp.Data.CheckoutURL
	if url == "" {
		url = resp.Data.UniversalURL
	}
	if url == "" {
		url = resp.Data.Deeplink
	}
	return resp.Data.PrepayID, url, nil
}

func (c *BinancePayClient) post(path string, body []byte, out interface{}) error {
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())
	nonce := randomNonce(32)
	payload := ts + "\n" + nonce + "\n" + string(body) + "\n"
	mac := hmac.New(sha512.New, []byte(c.SecretKey))
	mac.Write([]byte(payload))
	sig := strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))

	req, err := http.NewRequest(http.MethodPost, binancePayBase+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("BinancePay-Timestamp", ts)
	req.Header.Set("BinancePay-Nonce", nonce)
	req.Header.Set("BinancePay-Certificate-SN", c.CertSN)
	req.Header.Set("BinancePay-Signature", sig)

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return fmt.Errorf("binance pay http %d: %s", res.StatusCode, truncate(string(data), 300))
	}
	return json.Unmarshal(data, out)
}

func randomNonce(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
