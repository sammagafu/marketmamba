# Automated Trading Guide

This guide explains how to use the automated position opening and closing features.

## Overview

The bot now includes:
1. **Automated Signal Generation** - Generates trading signals based on technical analysis
2. **Position Monitor** - Continuously monitors positions for Take Profit (TP) and Stop Loss (SL)
3. **Automated Execution** - Executes signals while respecting all risk management rules
4. **Manual Override** - You can pause/resume automation at any time

## Key Features

### ✅ Automated Position Opening
- Generates signals every 10 seconds
- Analyzes trends using EMA (20, 50, 200)
- Validates spread and volatility
- Checks RSI for overbought/oversold conditions
- Respects all risk limits (max trades, daily loss, etc.)
- Calculates lot size automatically based on risk percentage

### ✅ Automated Position Closing
- Monitors positions every 5 seconds
- Closes at Take Profit automatically
- Closes at Stop Loss automatically
- Updates profit/loss and daily statistics
- Logs all closures for audit trail

### ✅ Risk Management
- Validates every signal against risk rules
- Prevents trading if paused
- Stops trading if daily loss limit hit
- Enforces max concurrent open trades
- Enforces max trades per day
- Requires minimum risk-reward ratio

## Starting Automation

### Command: `/autostart`

Enables automated trading. The bot will:
1. Generate trading signals automatically
2. Execute trades when conditions align
3. Close positions at TP/SL

```
User: /autostart
Bot: 🤖 *Automated trading started*

The bot will now automatically:
• Generate trading signals
• Open trades when conditions align
• Close positions at TP/SL
• Respect risk management rules

You can still use /pause to pause manually.
```

## Stopping Automation

### Command: `/autostop`

Disables automated trading but keeps manual control.

```
User: /autostop
Bot: ⏹️ *Automated trading stopped*

Manual /open and /close commands still work.
```

## Checking Automation Status

### Command: `/autostatus`

Shows the current automation state.

```
User: /autostatus
Bot: *Automation Status*
Automated Trading: ✅ Enabled
Trading Status: ✅ Trading Active
Daily Loss Hit: false
Last Active: 2026-05-21 22:30:15
```

## How Signals are Generated

### Technical Indicators Used

1. **EMA (Exponential Moving Average)**
   - EMA 20: Short-term trend
   - EMA 50: Medium-term trend
   - EMA 200: Long-term trend

2. **RSI (Relative Strength Index)**
   - Identifies overbought (>75) / oversold (<25)
   - Scalping avoids extremes for safety

3. **ATR (Average True Range)**
   - Measures volatility
   - Used to set Stop Loss and Take Profit
   - Filters out low-volatility periods

4. **Spread Filter**
   - Rejects trades if spread > 3 pips
   - Ensures profitability is possible

### Signal Types

#### 1. UPTREND_SCALP (Buy)
```
Conditions:
- Price at/near EMA20 (within 1%)
- EMA20 > EMA50 > EMA200 (strong uptrend)
- RSI < 70 (not overbought)
- Strength: 0.90 (very strong)

Entry: At current price
Stop Loss: 1 ATR below entry
Take Profit: 2 ATR above entry (1:2 risk-reward)
```

#### 2. DOWNTREND_SCALP (Sell)
```
Conditions:
- Price at/near EMA20 (within 1%)
- EMA20 < EMA50 < EMA200 (strong downtrend)
- RSI > 30 (not oversold)
- Strength: 0.90 (very strong)

Entry: At current price
Stop Loss: 1 ATR above entry
Take Profit: 2 ATR below entry (1:2 risk-reward)
```

#### 3. UPTREND (Buy)
```
Conditions:
- Price >= EMA50 (within 0.5%)
- EMA20 > EMA50 (moderate uptrend)
- RSI < 70 (not overbought)
- Strength: 0.70 (moderate)

Entry: At current price
Stop Loss: 1 ATR below entry
Take Profit: 2 ATR above entry (1:2 risk-reward)
```

#### 4. DOWNTREND (Sell)
```
Conditions:
- Price <= EMA50 (within 0.5%)
- EMA20 < EMA50 (moderate downtrend)
- RSI > 30 (not oversold)
- Strength: 0.70 (moderate)

Entry: At current price
Stop Loss: 1 ATR above entry
Take Profit: 2 ATR below entry (1:2 risk-reward)
```

## Control Flow

### What Happens When You Send `/autostart`

```
User sends /autostart
    ↓
Check if already running
    ↓
Check if paused
    ↓
Check if daily loss hit
    ↓
Enable automation
    ↓
Start signal monitor (every 10s)
Start position monitor (every 5s)
```

### What Happens Every 10 Seconds (Signal Generation)

```
Check bot state
    ↓
Is bot paused? → Return (do nothing)
    ↓
Is daily loss hit? → Return (do nothing)
    ↓
Generate signal from market data
    ↓
Is signal strong enough? → Continue
    ↓
Validate against risk rules
    ↓
Calculate lot size
    ↓
Execute trade
    ↓
Log trade to database
```

### What Happens Every 5 Seconds (Position Monitoring)

```
Get all open positions
    ↓
For each position:
    │
    ├→ Check if Take Profit hit
    │  ├→ YES: Close position, log profit, update stats
    │  └→ NO: Continue
    │
    └→ Check if Stop Loss hit
       ├→ YES: Close position, log loss, update stats
       └→ NO: Continue
```

## Risk Management Rules

All automated trades must pass these checks:

1. **Bot Not Paused**
   - Trading must not be manually paused

2. **Daily Loss Limit**
   - Total daily losses must not exceed X% of balance
   - Default: 2%

3. **Max Concurrent Trades**
   - Maximum number of positions open at once
   - Default: 2 trades

4. **Max Trades Per Day**
   - Maximum number of trades opened in a day
   - Default: 10 trades

5. **Risk-Reward Ratio**
   - Reward must be at least 1:1 of risk
   - Default: 1.0 (TP must be as far as SL)

6. **Spread Filter**
   - Spread must be < 3 pips
   - Protects against slippage

7. **Volatility Filter**
   - ATR must be > 0.05% of price
   - Prevents trading in dead zones

8. **RSI Filter**
   - RSI must be between 20-80
   - Avoids extreme overbought/oversold

## Example: Automated Trade Lifecycle

### 1. Signal Generation (10s interval)
```
Time: 22:30:05
EMA20: 1.1050
EMA50: 1.1040
EMA200: 1.1020
ATR: 0.0035
RSI: 60

Analysis: EMA20 > EMA50 > EMA200 = UPTREND
Signal Generated: BUY EURUSD (strength: 0.90)
```

### 2. Validation
```
Check rules:
✓ Bot not paused
✓ Daily loss: $-15 (within -$200 limit)
✓ Open trades: 1 (limit 2)
✓ Today trades: 5 (limit 10)
✓ Spread: 1.5 pips (< 3)
✓ Volatility: OK
✓ RSI: 60 (20-80 range)

Risk: Entry 1.1055 - SL 1.1020 = 35 pips = $350
Reward: TP 1.1090 - Entry 1.1055 = 35 pips = $350
Ratio: 1.0 ✓

Lot Size: $10,000 * 0.5% / 35 pips = 1.43 lots
```

### 3. Trade Execution
```
Command: OpenMarketOrder(EURUSD, BUY, 1.43, 1.1020, 1.1090)
Result: Position opened at 1.1055
Trade ID: trade_abc123
Position ID: pos_xyz789

Log:
- Trade record created
- Position record created
- Command logged
```

### 4. Position Monitoring (every 5s)

```
Check at 22:30:08:
Current: 1.1062
Entry: 1.1055
Profit: +0.0007 = +$7 (not TP yet)

Check at 22:30:13:
Current: 1.1090
Entry: 1.1055
Profit: +0.0035 = +$50

TP hit! Close position
- Status: CLOSED
- Exit Price: 1.1090
- Reason: TP
- Profit: +$50

Update daily stats:
- Trade count: 6
- Win count: 4
- Loss count: 2
- Total profit: $150
- Win rate: 66.7%
```

## Configuration

Edit `.env` to control automation behavior:

```env
# Risk limits (automation respects these)
MAX_RISK_PER_TRADE=0.005       # 0.5% per trade
MAX_DAILY_LOSS=0.02             # 2% daily max
MAX_OPEN_TRADES=2               # Max 2 concurrent
MAX_TRADES_PER_DAY=10           # Max 10/day
RISK_REWARD_RATIO=1.0           # Min 1:1

# Automation timing (in code, change monitor intervals)
SIGNAL_CHECK_INTERVAL=10s       # How often to generate signals
POSITION_CHECK_INTERVAL=5s      # How often to check TP/SL
```

## Testing Automation

### Step 1: Start Bot
```bash
docker-compose up -d
```

### Step 2: Enable Automation
```
/autostart
```

### Step 3: Monitor in Real-Time
```
/autostatus          # Check automation status
/positions           # See open positions
/balance             # Check account
/dailyreport         # See daily stats
```

### Step 4: Watch Positions Close
- The bot closes positions at TP automatically
- Log shows which closed and why (TP/SL)
- Daily stats update automatically

### Step 5: View Results
```
/dailyreport
Shows:
- Trades opened: X
- Wins: Y
- Losses: Z
- Win rate: %
- Net profit: $
```

## Manual Control During Automation

Even with automation enabled, you can:

1. **Pause Trading**
   ```
   /pause              # Stops all trading (auto and manual)
   /resume             # Resumes trading
   ```

2. **Close Specific Position**
   ```
   /close <position_id>  # Close one position manually
   ```

3. **Close All**
   ```
   /closeall           # Close all positions (auto stops)
   ```

4. **Stop Automation**
   ```
   /autostop           # Stop automation (manual still works)
   ```

## Safety Mechanism: Daily Loss Limit

When daily loss exceeds the limit:

```
Loss Limit: 2% of $10k = $200
Current Loss: $201 (2.01%)
    ↓
Daily Loss Limit Hit!
    ↓
Bot pauses automatically
    ↓
/autostop (automation disabled)
    ↓
Manual trading disabled too
    ↓
User must /resume to continue
```

When you `/resume`:
```
Bot resets:
- DailyLossHit: false
- IsPaused: false
- AutoTradingActive: false (must /autostart again)

Ready to trade again
```

## Monitoring & Logs

### Database Tables Updated

During automation, these tables are updated:

- **trades** - New trade record for each opened position
- **positions** - New position record, updated with profit/loss
- **daily_stats** - Win/loss count, profit, daily stats
- **command_logs** - "AUTO_TRADE" and "AUTO_CLOSE" entries
- **bot_states** - Last active time updated

### Log Output

```
[INFO] Starting signal monitor for user 123456789
[INFO] Starting position monitor for user 123456789
[INFO] Signal generated for user 123456789: EURUSD BUY
[INFO] Trade executed for user 123456789: EURUSD BUY
[INFO] Position closed for user 123456789: EURUSD @ 1.10900
```

## Troubleshooting

### Automation Not Generating Signals

1. Check status:
   ```
   /autostatus
   ```

2. Check logs:
   ```
   docker-compose logs -f app | grep -i signal
   ```

3. Verify:
   - Bot is not paused (`/resume`)
   - Daily loss limit not hit (`/dailyreport`)
   - Risk settings configured (`/risk`)

### Positions Not Closing at TP/SL

1. Check positions:
   ```
   /positions
   ```

2. Verify position monitor is running:
   ```
   docker-compose logs app | grep "position monitor"
   ```

3. Check if paused:
   ```
   /autostatus
   ```

### High Slippage or Unexpected Entries

1. Current implementation uses simulated prices
2. When connecting real broker, slippage will be real
3. Adjust spread filter in signal_generator.go if needed

### Bot Paused After Daily Loss

1. Check daily stats:
   ```
   /dailyreport
   ```

2. If negative profit < daily loss limit:
   ```
   /resume
   ```

3. Daily loss limit resets at midnight

## Advanced Configuration

### Changing Signal Generation Interval

Edit `cmd/server/main.go`:
```go
// Current: 10 seconds
sigMonitor := trading.NewSignalMonitor(sigGen, executor, db, primaryUserID, 10*time.Second)

// Change to 5 seconds for more frequent signals
sigMonitor := trading.NewSignalMonitor(sigGen, executor, db, primaryUserID, 5*time.Second)
```

### Changing Position Check Interval

Edit `cmd/server/main.go`:
```go
// Current: 5 seconds
posMonitor := trading.NewPositionMonitor(b, db, primaryUserID, 5*time.Second)

// Change to 2 seconds for faster TP/SL detection
posMonitor := trading.NewPositionMonitor(b, db, primaryUserID, 2*time.Second)
```

### Adjusting Signal Strength Threshold

Edit `cmd/server/main.go`:
```go
// Current: minimum 0.7 strength
sigGen := trading.NewSignalGenerator("EURUSD", 0.7, cfg.Risk.RiskRewardRatio)

// Higher threshold = fewer but stronger signals
sigGen := trading.NewSignalGenerator("EURUSD", 0.85, cfg.Risk.RiskRewardRatio)

// Lower threshold = more signals (higher risk)
sigGen := trading.NewSignalGenerator("EURUSD", 0.5, cfg.Risk.RiskRewardRatio)
```

### Adding Multiple Symbols

Currently monitors EURUSD. To add more:

```go
symbols := []string{"EURUSD", "GBPUSD", "USDJPY"}
for _, symbol := range symbols {
    sigGen := trading.NewSignalGenerator(symbol, 0.7, cfg.Risk.RiskRewardRatio)
    sigMonitor := trading.NewSignalMonitor(sigGen, executor, db, primaryUserID, 10*time.Second)
    sigMonitor.Start(ctx)
}
```

## Important Notes

1. **Mock Trading Only**
   - Currently uses simulated market data
   - Real prices come from broker API integration
   - Test thoroughly before using real account

2. **No Slippage in Mock Mode**
   - Real trading will have slippage costs
   - Adjust TP/SL to account for this

3. **Continuous Monitoring**
   - Position monitor runs every 5 seconds
   - Checks for TP/SL hits
   - Closes positions automatically

4. **Audit Trail**
   - Every trade logged to database
   - Command logs show AUTO_TRADE and AUTO_CLOSE
   - All decisions and prices recorded

5. **Daily Reset**
   - Daily stats reset at midnight
   - Daily loss limit resets
   - Trade count resets

## Next Steps

1. Test automation thoroughly with mock broker
2. Monitor logs and daily reports
3. Adjust risk settings and intervals
4. When ready: Integrate real broker API
5. Use demo account first
6. Start with very small position sizes

**Remember: Always test with mock data first!**
