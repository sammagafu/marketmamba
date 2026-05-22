# Market Mamba Web Dashboard

## What you get

- **URL:** `https://marketmamba.kkooapp.co.tz` (via nginx → `127.0.0.1:8090`)
- **UI:** Status, balance, positions, broker picker
- **Brokers:** Mock works today; OANDA / MetaAPI / Alpaca / Custom are listed for future adapters

## VPS setup

### 1. Run DB migration (once)

```bash
docker compose -p marketmamba exec postgres psql -U forexbot -d forexbot -f /docker-entrypoint-initdb.d/../migrations/002_broker_connections.sql
```

Or from host:

```bash
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/002_broker_connections.sql
```

### 2. Add to `.env`

```env
HTTP_PORT=8090
ENABLE_WEB=true
WEB_API_KEY=your_long_random_secret
CORS_ORIGINS=https://marketmamba.kkooapp.co.tz
BROKER_ENCRYPTION_KEY=your_32_char_encryption_secret
```

### 3. Redeploy

```bash
git pull
docker compose -p marketmamba up -d --build
```

### 4. Nginx

Copy `deploy/nginx-marketmamba.conf.example` to nginx sites-enabled, then:

```bash
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d marketmamba.kkooapp.co.tz
```

### 5. Open the site

Enter `WEB_API_KEY` in the dashboard header, then connect **Mock (Demo)** broker.

## Adding a real broker

1. Implement `Broker` interface in `internal/broker/<name>.go`
2. Register in `internal/broker/factory.go` and set `Status: "live"` in `registry.go`
3. Redeploy

For MT4/MT5 without custom code, plan for **MetaAPI** or similar bridge.

## “Any broker” realistically

| Approach | Effort |
|----------|--------|
| One adapter per broker (OANDA, Alpaca, …) | Medium per broker |
| MetaAPI / universal bridge | Paid SaaS, one integration |
| Custom REST adapter | You host a small bridge service |

The UI already supports saving credentials; trading goes live when the Go adapter exists.
