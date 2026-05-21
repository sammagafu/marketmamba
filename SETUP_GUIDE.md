# Setup Guide - Forex Scalping Bot

Step-by-step guide to get your bot running in 10 minutes.

## Step 1: Get Telegram Bot Token

1. Open Telegram and search for **@BotFather**
2. Send message: `/newbot`
3. Follow the prompts:
   - **Name:** Forex Scalping Bot (or your choice)
   - **Username:** forex_scalping_bot_<random> (must be unique)
4. BotFather will reply with your **API Token** - copy this
5. Example token format: `123456789:ABCdefGHIjklmNOPqrsTUVwxyz1234567`

## Step 2: Get Your Telegram User ID

1. Open Telegram and search for **@userinfobot**
2. Send message: `/start`
3. Bot will reply with your **User ID** - copy this
4. Example: `123456789`

## Step 3: Clone and Configure

```bash
# Clone or download the project
cd forex-bot

# Copy environment template
cp .env.example .env

# Edit .env with your values
nano .env
```

Update `.env`:
```env
TELEGRAM_BOT_TOKEN=<your_token_from_step_1>
TELEGRAM_ALLOWED_USER_IDS=<your_user_id_from_step_2>
DATABASE_URL=postgres://forexbot:forexbot_password_change_me@localhost:5432/forexbot?sslmode=disable
BROKER_PROVIDER=mock
MAX_RISK_PER_TRADE=0.005
MAX_DAILY_LOSS=0.02
MAX_OPEN_TRADES=2
MAX_TRADES_PER_DAY=10
APP_ENV=development
PORT=8080
```

## Step 4: Run with Docker (Recommended)

```bash
# Start services
docker-compose up -d

# Check if running
docker-compose ps

# View logs
docker-compose logs -f app
```

Expected output:
```
[INFO] Starting Forex Scalping Bot
[INFO] Database connected successfully
[INFO] Bot initialized successfully
[INFO] Telegram bot started: @<your_bot_username>
```

## Step 5: Test Your Bot

1. Open Telegram and find your bot (or use the link from BotFather)
2. Send `/start` - you should see the help menu
3. Send `/balance` - you should see mock account balance
4. Send `/positions` - you should see "No open positions"

## Step 6: Test Opening a Trade

Send this command:
```
/open EURUSD BUY 1.0 1.0900 1.1000
```

Expected response:
```
✅ Trade Opened
Symbol: EURUSD
Type: BUY
Entry: 1.10000
Stop Loss: 1.09000
Take Profit: 1.10000
Position ID: mock_pos_1
```

## Step 7: Monitor Positions

Send `/positions` to see your open trade.

## Available Commands Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `/start` | Show help | `/start` |
| `/status` | Bot status | `/status` |
| `/balance` | Account balance | `/balance` |
| `/positions` | List open trades | `/positions` |
| `/open` | Open trade | `/open EURUSD BUY 1.0 1.0900 1.1000` |
| `/close` | Close trade | `/close mock_pos_1` |
| `/closeall` | Close all trades | `/closeall` |
| `/pause` | Pause trading | `/pause` |
| `/resume` | Resume trading | `/resume` |
| `/risk` | Risk settings | `/risk` |
| `/dailyreport` | Daily stats | `/dailyreport` |

## Troubleshooting

### "❌ Unauthorized access"
- Check your Telegram User ID is correct in `.env`
- Restart the bot after changing `.env`

### "❌ Error fetching balance"
- Check Docker logs: `docker-compose logs app`
- Verify database is running: `docker-compose ps postgres`

### "❌ Unknown command"
- Make sure command starts with `/`
- Verify bot is running (check logs)

### Can't connect to bot
- Double-check bot token in `.env`
- Verify you have internet connection
- Restart: `docker-compose restart app`

### Database connection error

```bash
# Check database status
docker-compose ps postgres

# View database logs
docker-compose logs postgres

# Recreate database
docker-compose down
docker-compose up -d
```

## Next Steps

### 1. Test More Thoroughly
- Try different symbols (GBPUSD, USDJPY, etc.)
- Test pause/resume functionality
- Check daily report

### 2. Configure Risk Settings
Edit `.env` to customize:
```env
MAX_RISK_PER_TRADE=0.01        # 1% per trade
MAX_DAILY_LOSS=0.05             # 5% daily limit
MAX_OPEN_TRADES=5               # Max 5 concurrent
MAX_TRADES_PER_DAY=20           # Max 20/day
```

### 3. Connect Real Broker (Future)
When ready to use a real broker:
1. Create a new broker implementation (e.g., OANDA)
2. Implement the `Broker` interface
3. Update `cmd/server/main.go` to use it
4. Test thoroughly with demo account first

### 4. Deploy to VPS
See `README.md` for VPS deployment instructions

## Security Checklist

- [ ] Changed database password from default
- [ ] .env file is NOT in version control
- [ ] Using only your Telegram User ID
- [ ] Using mock broker (not production account)
- [ ] Have daily loss limits configured
- [ ] Have max open trades limit set
- [ ] Server has firewall configured

## Performance Tips

1. **Monitor Resource Usage:**
   ```bash
   docker-compose stats
   ```

2. **View Application Logs:**
   ```bash
   docker-compose logs -f app --tail 50
   ```

3. **Backup Database:**
   ```bash
   docker-compose exec postgres pg_dump -U forexbot forexbot > backup.sql
   ```

## Getting Help

- Check logs: `docker-compose logs -f`
- Review README.md for more details
- Check .env.example for all available options
- Verify Telegram bot token format: `XXXXXXXXX:XXXXXXXXXXXXXXXXXXXXXXXXXXX`

## Quick Commands Reference

```bash
# View all containers
docker-compose ps

# View live logs
docker-compose logs -f app

# Restart bot
docker-compose restart app

# Stop everything
docker-compose down

# View database
docker-compose exec postgres psql -U forexbot -d forexbot

# List trades
docker-compose exec postgres psql -U forexbot -d forexbot -c "SELECT * FROM trades LIMIT 5;"
```

## Success!

If you can:
- ✅ Start the bot
- ✅ Receive `/start` command response
- ✅ Open and close mock trades
- ✅ View balance and positions

**You're ready to explore further!** 🚀

Next: Read the full README.md for VPS deployment and real broker integration.
