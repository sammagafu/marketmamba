# VPS deployment — Market Mamba

Step-by-step guide for deploying on a Linux VPS (e.g. alongside kkoo on `marketmamba.kkooapp.co.tz`).

---

## 1. DNS

Add an **A record**:

| Host | Value |
|------|--------|
| `marketmamba` | Your VPS public IP |

---

## 2. Clone and create `.env`

```bash
ssh sammy@kkooapp.co.tz
cd /home/sammy/marketmamba
# first time only:
# git clone git@github.com:sammagafu/marketmamba.git /home/sammy/marketmamba
cp .env.example .env
nano .env
```

**Copy `.env` from your Mac:**

```bash
scp /Users/codexl-008/iloveprojects/forex-bot/.env \
  sammy@kkooapp.co.tz:/home/sammy/marketmamba/.env
```

### Required variables

| Variable | What to put |
|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | From @BotFather |
| `TELEGRAM_BOT_USERNAME` | `market_mamba_bot` |
| `TELEGRAM_ADMIN_USER_IDS` | Your Telegram user ID (e.g. `5311857635`) |
| `TELEGRAM_ALLOWED_USER_IDS` | Same or comma-separated list |
| `WEB_API_KEY` | Random secret — `openssl rand -hex 32` |
| `WEB_SESSION_SECRET` | Random secret — `openssl rand -hex 32` |
| `BROKER_ENCRYPTION_KEY` | Random secret (32+ chars) |
| `CORS_ORIGINS` | `https://marketmamba.kkooapp.co.tz` |

### Admin email login (optional, recommended)

| Variable | What to put |
|----------|-------------|
| `ADMIN_EMAIL` | Your email (e.g. `iammagafu@gmail.com`) |
| `ADMIN_PASSWORD` | Strong password (8+ chars) — **VPS only, never commit** |
| `ADMIN_TELEGRAM_ID` | Same as `TELEGRAM_ADMIN_USER_IDS` |

Example block in `.env` (replace password locally on the server):

```env
ADMIN_EMAIL=iammagafu@gmail.com
ADMIN_PASSWORD=your-strong-password-here
ADMIN_TELEGRAM_ID=5311857635
```

---

## 3. Start Docker

```bash
docker compose -p marketmamba up -d --build
```

Check:

```bash
docker compose -p marketmamba ps
curl -s http://127.0.0.1:8090/health
```

---

## 4. Database migrations

Run once (skip lines that error with “already exists”):

```bash
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/002_broker_connections.sql
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/003_users_subscriptions.sql
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/004_web_admins.sql
```

Or run the fix script (stops old containers, rebuilds, migrates):

```bash
bash scripts/vps-fix.sh
```

---

## 5. Create admin user (email + password)

```bash
docker compose -p marketmamba exec app ./seedadmin
```

You should see: `Web admin ready — log in with ADMIN_EMAIL on the dashboard`

Log in at `https://marketmamba.kkooapp.co.tz` → **Admin login (email)**.

---

## 6. Nginx + SSL (HTTPS)

DNS must point `marketmamba.kkooapp.co.tz` → this server before running certbot.

**Automated (recommended):**

```bash
cd /home/sammy/marketmamba
sudo bash scripts/setup-ssl.sh
```

**Manual:**

```bash
# 1) HTTP first (Let's Encrypt needs port 80)
sudo cp deploy/nginx-marketmamba-http.conf.example \
  /etc/nginx/sites-available/marketmamba.kkooapp.co.tz
sudo ln -sf /etc/nginx/sites-available/marketmamba.kkooapp.co.tz /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# 2) Get certificate (enter email when prompted)
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d marketmamba.kkooapp.co.tz

# 3) Full SSL config (HTTP redirect + Telegram popup header)
sudo cp deploy/nginx-marketmamba.conf.example \
  /etc/nginx/sites-available/marketmamba.kkooapp.co.tz
sudo nginx -t && sudo systemctl reload nginx
```

**Verify:**

```bash
curl -sI https://marketmamba.kkooapp.co.tz/health
sudo certbot renew --dry-run
```

---

## 7. Telegram Login (OIDC)

Per [Telegram Login docs](https://core.telegram.org/bots/telegram-login):

1. @BotFather → **Bot Settings → Web Login**
2. **Allowed URLs:** `https://marketmamba.kkooapp.co.tz` (add `http://localhost:5173` for local dev if needed)
3. Note **Client ID** (`8040019896`) — must match `TELEGRAM_BOT_CLIENT_ID` in `.env`
4. Site uses `oauth.telegram.org/js/telegram-login.js` + server verifies `id_token` (JWKS)

---

## 8. Stop local bot (important)

Only **one** instance per bot token:

```bash
# On your Mac
docker compose -p marketmamba down
```

---

## Useful commands

```bash
# Update after git pull
cd /home/sammy/marketmamba && git pull
docker compose -p marketmamba up -d --build

# Logs
docker compose -p marketmamba logs -f app

# Restart
docker compose -p marketmamba restart app

# Re-seed admin after password change
docker compose -p marketmamba exec app ./seedadmin

# Shell into DB
docker compose -p marketmamba exec postgres psql -U forexbot -d forexbot
```

---

## Admin features (web)

After email admin login:

- Dashboard stats (users, subs, auto-trading)
- Activate subscription (manual payment)
- Block / unblock user
- Revoke active subscription

## Admin features (Telegram)

- `/admin stats`
- `/admin activate <telegram_id> <days>`

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `go.mod requires go >= 1.25` | `git pull` — Dockerfile uses Go 1.23; rebuild |
| `bind :8090: address already in use` | `docker compose -p marketmamba down` then `bash scripts/vps-fix.sh` |
| `relation "broker_connections" does not exist` | Run migration `002_broker_connections.sql` |
| `Admins: []` in logs | Fix `.env`: `TELEGRAM_ADMIN_USER_IDS=5311857635`, restart app |
| `WEB_API_KEY variable is not set` | Create `.env` in project folder, restart compose |
| `getUpdates conflict` | Stop bot on Mac; never run `docker compose exec app ./server` (starts 2nd bot). Use `./seedadmin` only |
| `bind :8090 already in use` | You started a second `./server` in the container; use `./seedadmin` for admin seed |
| 502 from nginx | `docker compose -p marketmamba ps`, check port 8090 |
| `GET /` 404 in browser | Rebuild app (`git pull && docker compose -p marketmamba up -d --build`); test `curl http://127.0.0.1:8090/` on VPS; fix nginx `proxy_pass` to **8090** |
| Email login fails | Run migration 004 + `seed-admin`; check `TELEGRAM_ADMIN_USER_IDS` |
| Widget missing | BotFather Web Login allowed URLs, use HTTPS |
