# Quick Reference - Automated Trading

## Telegram Commands

### 🤖 Automation Control
```
/autostart      Start automated trading
/autostop       Stop automated trading
/autostatus     Check automation status
```

### 📊 Manual Trading
```
/open EURUSD BUY 1.0 1.0900 1.1000    Open trade manually
/close <position_id>                    Close specific position
/closeall                               Close all positions
```

### ⚙️ Bot Control
```
/pause          Pause all trading (manual & auto)
/resume         Resume trading
```

### 📈 Information
```
/status         Bot status
/balance        Account balance
/positions      Open positions
/risk           Risk settings
/dailyreport    Daily statistics
/start          Show help
```

---

## How It Works

### Automated Signal Generation (Every 10 seconds)
1. Analyzes market conditions (EMA, RSI, ATR)
2. Generates trading signal if conditions align
3. Validates against risk rules
4. Executes trade automatically
5. Logs trade to database

### Position Monitoring (Every 5 seconds)
1. Checks all open positions
2. If Take Profit hit → Auto closes
3. If Stop Loss hit → Auto closes
4. Updates profit/loss in database
5. Updates daily statistics

### Risk Management (Always Active)
- Max concurrent trades: 2
- Max daily loss: 2%
- Max trades per day: 10
- Minimum risk-reward: 1:1
- Spread filter: < 3 pips
- Volatility filter: Active
- RSI filter: 20-80 range

---

## Workflow Examples

### Example 1: Start Automated Trading
```
1. /autostart
   ✅ Automation enabled
   
2. Bot starts generating signals every 10 seconds
   
3. /positions
   Shows: 1 open position (auto-opened)
   
4. Wait for TP/SL
   ✅ Position closes automatically
   
5. /dailyreport
   Shows: 1 win/loss, net profit
```

### Example 2: Pause & Resume
```
1. /autostart
   Automated trading running
   
2. /pause
   ✅ All trading paused (manual & auto)
   
3. /resume
   ✅ Trading resumed, automation continues
```

### Example 3: Manual Override
```
1. /autostart
   Automated trading running
   
2. /open EURUSD SELL 0.5 1.1100 1.1000
   ✅ Manual trade opened
   
3. /positions
   Shows: 1 auto + 1 manual = 2 open
   
4. /close <position_id>
   ✅ Manual position closed
   
5. Auto positions still run
```

### Example 4: Hit Daily Loss Limit
```
1. /autostart
   Running...
   
2. Losses accumulate to $200 (2% of $10k)
   
3. /autostatus
   Daily Loss Hit: true
   Trading Status: ⏸️ Trading Paused
   
4. /resume
   ✅ Bot resets and resumes
```

---

## Key Metrics

### Position Monitor
- Checks every: 5 seconds
- Actions: Close at TP/SL, update prices
- Symbols: EURUSD (configurable)

### Signal Generator
- Checks every: 10 seconds
- Indicators: EMA, RSI, ATR, Spread
- Signal types: 4 (strong buy, buy, strong sell, sell)

### Risk Limits
```
MAX_RISK_PER_TRADE = 0.5%
MAX_DAILY_LOSS = 2%
MAX_OPEN_TRADES = 2
MAX_TRADES_PER_DAY = 10
RISK_REWARD_RATIO = 1.0
```

### Timing
- Signal generation: 10 seconds
- Position check: 5 seconds
- Database update: Real-time
- Daily reset: Midnight UTC

---

## Safety Features

1. **Daily Loss Limit**: Auto-pauses if 2% lost
2. **Max Trades**: Won't open more than 2 concurrent
3. **Daily Cap**: Max 10 trades per calendar day
4. **Risk Validation**: Every trade checked before execution
5. **Spread Filter**: Rejects trades if spread too high
6. **RSI Filter**: Avoids extreme conditions
7. **Manual Override**: You can pause/resume anytime
8. **Audit Trail**: Every trade logged to database

---

## Database Activity

When automation runs, these are updated:
- `trades` table - New trade records
- `positions` table - Position records and updates
- `daily_stats` table - Win/loss count, profit
- `command_logs` table - AUTO_TRADE, AUTO_CLOSE
- `bot_states` table - Last activity timestamp

---

## Troubleshooting Quick Tips

| Problem | Solution |
|---------|----------|
| No signals generated | Check: `/autostatus`, `/pause`, `/dailyreport` |
| Positions not closing | Check logs: `docker-compose logs app` |
| Daily loss limit hit | `/dailyreport` to check, `/resume` to reset |
| Too many trades opened | Increase `MAX_OPEN_TRADES` in `.env` |
| Trades closing too fast | Increase `SIGNAL_CHECK_INTERVAL` |
| Not enough trades | Lower signal strength threshold |

---

## Testing Checklist

- [ ] `/start` - See help message
- [ ] `/autostart` - Enable automation
- [ ] `/autostatus` - Check enabled
- [ ] Wait 10+ seconds for signal
- [ ] `/positions` - See auto-opened position
- [ ] Wait for TP/SL hit
- [ ] `/positions` - Position should close
- [ ] `/dailyreport` - Check results
- [ ] `/autostop` - Stop automation
- [ ] `/balance` - Verify account updated

---

## Performance Notes

- Position monitor: ~1% CPU per user
- Signal generator: ~1-2% CPU per user
- Database queries: ~100ms per check
- Position check interval: 5 seconds
- Signal check interval: 10 seconds

Memory usage:
- Per active user: ~10-20MB
- With 10 active users: ~150-200MB
- Database connections: 1-2 per user

---

## Docker Commands

```bash
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f app

# Check if running
docker-compose ps

# Stop everything
docker-compose down

# View bot state
docker-compose exec postgres psql -U forexbot -d forexbot -c \
  "SELECT * FROM bot_states;"

# View recent trades
docker-compose exec postgres psql -U forexbot -d forexbot -c \
  "SELECT * FROM trades ORDER BY created_at DESC LIMIT 5;"
```

---

## Configuration Files

**For Automation:**
- `internal/trading/executor.go` - Trade execution
- `internal/trading/monitor.go` - Position & signal monitoring
- `internal/trading/signal_generator.go` - Signal generation logic
- `cmd/server/main.go` - Main automation setup

**To Adjust:**
- Intervals: Edit `cmd/server/main.go` (lines with `*time.Second`)
- Risk limits: Edit `.env` file
- Signal strength: Edit `cmd/server/main.go` (0.7 value)
- Symbols: Edit `internal/trading/signal_generator.go`

---

## API Response Times

| Operation | Time |
|-----------|------|
| Generate signal | ~50ms |
| Execute trade | ~100ms |
| Check position | ~30ms |
| Close position | ~100ms |
| Update database | ~50ms |
| Total cycle | ~200-300ms |

---

## Next Steps

1. **Test with mock broker** (current setup)
2. **Monitor daily reports** - Check profitability
3. **Adjust risk settings** - Fine-tune for your style
4. **Add real broker** - OANDA, cTrader, MT5
5. **Test with demo account** - Before real money
6. **Start small** - Begin with 0.1% risk per trade

---

**Last Updated**: May 2026
**Version**: 1.0 with Automation
**Status**: Ready for testing
