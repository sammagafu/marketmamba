# Market Mamba

A production-ready Go backend for automated forex trading with Telegram bot control, risk management, and broker integration.

## ⚠️ Risk Disclaimer

**This bot is for educational and testing purposes only.** Forex trading carries substantial risk of loss. Do not use real account credentials without thorough testing. Always use stop losses and proper risk management. The developers are not responsible for trading losses.

## Features

### Core Functionality
- 🤖 Telegram bot integration with command control
- 📊 Real-time position monitoring
- 💰 Account balance and equity tracking
- 🛡️ Comprehensive risk management system
- 📱 Mobile-friendly Telegram commands
- 📈 Daily trading statistics
- 🔒 User ID-based access control
- 📝 Command audit logging

### Risk Management
- Max risk per trade percentage
- Daily loss limit enforcement
- Maximum concurrent open trades
- Daily trade count limits
- Risk-reward ratio validation
- Stop loss requirement enforcement
- Automatic trading pause on daily loss hit

### Architecture
- Clean architecture with dependency injection
- Interface-based broker abstraction (easy to integrate OANDA, cTrader, MT5)
- PostgreSQL storage layer
- Comprehensive logging
- Docker containerization
- Environment-based configuration

## Project Structure

```
forex-bot/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── broker/
│   │   └── broker.go              # Broker interface & mock implementation
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── logger/
│   │   └── logger.go              # Logging utilities
│   ├── models/
│   │   └── models.go              # Data models
│   ├── risk/
│   │   ├── risk.go                # Risk validation logic
│   │   └── risk_test.go           # Unit tests
│   ├── storage/
│   │   └── storage.go             # Database operations
│   ├── strategy/
│   │   └── scalping.go            # Strategy placeholders
│   ├── telegram/
│   │   └── telegram.go            # Telegram bot handler
│   └── utils/
│       └── utils.go               # Utility functions
├── migrations/
│   └── 001_init_schema.sql        # Database schema
├── .env.example                    # Environment variables template
├── Dockerfile                      # Container image
├── docker-compose.yml              # Container orchestration
├── go.mod & go.sum                # Go dependencies
└── README.md                       # This file
```

## Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Docker & Docker Compose (optional)
- Telegram Bot Token (from BotFather)

## Quick Start

### 1. Clone and Setup

```bash
cd forex-bot
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` with your settings:

```bash
# Get Telegram Bot Token from @BotFather on Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Your Telegram User ID (get from @userinfobot)
TELEGRAM_ALLOWED_USER_IDS=123456789

# Database URL
DATABASE_URL=postgres://user:password@localhost:5432/forexbot

# Risk settings (examples)
MAX_RISK_PER_TRADE=0.005        # 0.5% per trade
MAX_DAILY_LOSS=0.02              # 2% daily max loss
MAX_OPEN_TRADES=2                # Max 2 concurrent trades
MAX_TRADES_PER_DAY=10            # Max 10 trades/day
```

### 3. Docker Setup (Recommended)

```bash
# Start database and app
docker-compose up -d

# Check logs
docker-compose logs -f app

# Run migrations
docker-compose exec app psql -U forexbot -d forexbot -f migrations/001_init_schema.sql
```

### 4. Local Setup (Without Docker)

```bash
# Install dependencies
go mod download

# Create PostgreSQL database
createdb forexbot

# Run migrations
psql -U forexbot -d forexbot -f migrations/001_init_schema.sql

# Run the application
go run cmd/server/main.go
```

## Telegram Bot Commands

### Status & Information
- `/start` - Show help and available commands
- `/status` - Bot and trading status
- `/balance` - Account balance and equity
- `/positions` - List all open positions
- `/dailyreport` - Daily trading statistics
- `/risk` - Display risk settings

### Trade Management
- `/open EURUSD BUY 1.0 1.0900 1.1000` - Open trade
  - Format: `/open <SYMBOL> <BUY|SELL> <QUANTITY> <STOPLOSS> <TAKEPROFIT>`
- `/close <POSITION_ID>` - Close specific position
- `/closeall` - Close all open positions

### Bot Control
- `/pause` - Pause trading (manual control only)
- `/resume` - Resume trading

## Configuration Details

### Risk Management Settings

```go
MaxRiskPerTrade  = 0.005    // Risk max 0.5% per trade
MaxDailyLoss     = 0.02     // Stop trading after 2% daily loss
MaxOpenTrades    = 2        // Never have more than 2 open trades
MaxTradesPerDay  = 10       // Maximum 10 trades per calendar day
RiskRewardRatio  = 1.0      // Minimum reward must be >= risk
```

### Database Schema

The system stores:
- **trades** - Historical trade records
- **positions** - Open trading positions
- **accounts** - Account balance and equity
- **risk_settings** - User risk configuration
- **daily_stats** - Daily performance metrics
- **bot_states** - Bot pause/resume state
- **command_logs** - Audit trail of commands

## Broker Integration

### Current Implementation
- **Mock Broker** - For development and testing

### Adding a New Broker

1. Implement the `Broker` interface in `internal/broker/broker.go`:

```go
type Broker interface {
    GetBalance() (float64, error)
    GetEquity() (float64, error)
    GetOpenPositions() ([]*models.Position, error)
    OpenMarketOrder(symbol, orderType string, quantity, stopLoss, takeProfit float64) (*models.Position, error)
    ClosePosition(positionID string) error
    CloseAllPositions() error
    ModifyStopLoss(positionID string, newStopLoss float64) error
    ModifyTakeProfit(positionID string, newTakeProfit float64) error
}
```

2. Create a new broker package (e.g., `internal/broker/oanda.go`)

3. Update `cmd/server/main.go` to instantiate your broker:

```go
var b broker.Broker
if cfg.Broker.Provider == "oanda" {
    b, err = broker.NewOandaBroker(cfg.OandaConfig)
    // ...
}
```

## Testing

### Run Unit Tests
```bash
go test ./internal/risk -v
```

### Test Mock Trading
1. Message your bot with `/start`
2. Open a position: `/open EURUSD BUY 1.0 1.0900 1.1000`
3. View positions: `/positions`
4. Close position: `/close <POSITION_ID>`

## VPS Deployment

**Full guide:** [VPS_DEPLOY.md](./VPS_DEPLOY.md) · **Operator index:** [Agent.md](./Agent.md)

```bash
git clone git@github.com:sammagafu/marketmamba.git forex-bot
cd forex-bot
cp .env.example .env && nano .env   # TELEGRAM_BOT_TOKEN, WEB_API_KEY, secrets, ADMIN_EMAIL

make vps-up
make vps-migrate
make vps-seed-admin    # email admin login (ADMIN_EMAIL / ADMIN_PASSWORD in .env)

make vps-logs
```

Site: `https://marketmamba.kkooapp.co.tz` — nginx example in `deploy/nginx-marketmamba.conf.example`.

**Admin login:** email + password on the dashboard, or Telegram `/admin` commands (same Telegram ID in `TELEGRAM_ADMIN_USER_IDS`).

## Code Quality

### Project Features
- ✅ Error handling with context
- ✅ Structured logging
- ✅ Unit tests for risk module
- ✅ Interface-based design
- ✅ Configuration from environment
- ✅ SQL injection prevention (parameterized queries)
- ✅ Concurrent position tracking
- ✅ Comprehensive audit logging

### Build & Run
```bash
# Download dependencies
go mod download

# Build binary
go build -o forex-bot cmd/server/main.go

# Run with custom env file
source .env && ./forex-bot
```

## Future Enhancements

### Automated Trading
- [ ] Technical analysis integration (ATR, EMA, RSI)
- [ ] Scalping strategy implementation
- [ ] News filter integration
- [ ] Session-based trading rules
- [ ] Spread monitoring

### Broker Support
- [ ] OANDA REST API v20
- [ ] cTrader REST API
- [ ] MetaTrader 5 SDK
- [ ] IB (Interactive Brokers)

### Monitoring
- [ ] REST API for webhooks
- [ ] Discord bot notifications
- [ ] Slack integration
- [ ] Email alerts
- [ ] Grafana dashboards

### Advanced Features
- [ ] Portfolio optimization
- [ ] Multi-account support
- [ ] Strategy backtesting engine
- [ ] Machine learning predictions
- [ ] Trailing stop implementation

## Troubleshooting

### Database Connection Error
```bash
# Check PostgreSQL is running
docker-compose ps

# Check connection string in .env
DATABASE_URL=postgres://user:password@localhost:5432/forexbot
```

### Telegram Bot Not Responding
```bash
# Verify bot token
echo $TELEGRAM_BOT_TOKEN

# Check logs
docker-compose logs app | grep -i telegram

# Test token with curl
curl https://api.telegram.org/bot<TOKEN>/getMe
```

### High Memory Usage
- Reduce `MAX_OPEN_TRADES`
- Implement position cleanup jobs
- Check for goroutine leaks in logs

## Security Considerations

1. **Never commit .env file** - Use `.env.example` as template
2. **Use strong database passwords** - Change default in docker-compose
3. **Restrict Telegram access** - Only add your user ID(s)
4. **Enable HTTPS** - Use reverse proxy (nginx) in production
5. **Audit logs** - Regularly review command logs
6. **Update dependencies** - Run `go get -u` periodically
7. **Secure VPS** - Firewall, SSH key-only access, fail2ban
8. **No credential logging** - API keys never logged or stored

## Support & Contributing

For issues or contributions:
1. Check existing issues
2. Create detailed bug reports with logs
3. Include environment info (Go version, OS)
4. Test changes against mock broker

## License

This project is provided as-is for educational purposes.

## Additional Resources

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [Go Database/SQL](https://golang.org/pkg/database/sql/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [Forex Trading Best Practices](https://www.investopedia.com/articles/forex/)

---

**Remember:** Always start with small position sizes, test thoroughly with a demo account, and never risk more than you can afford to lose.
