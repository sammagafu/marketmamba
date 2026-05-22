package telegram

import (
	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"forex-bot/internal/logger"
)

// ConfigureMiniApp sets the bot menu button to open the web Mini App dashboard.
func (tb *TelegramBot) ConfigureMiniApp(appURL string) {
	if appURL == "" {
		return
	}
	btn := map[string]interface{}{
		"type": "web_app",
		"text": "📊 Dashboard",
		"web_app": map[string]string{
			"url": appURL,
		},
	}
	raw, _ := json.Marshal(btn)
	params := tgbotapi.Params{}
	params["menu_button"] = string(raw)
	_, err := tb.api.MakeRequest("setChatMenuButton", params)
	if err != nil {
		logger.Warn("setChatMenuButton: %v", err)
		return
	}
	logger.Info("Telegram Mini App menu → %s", appURL)
}
