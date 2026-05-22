# Market Mamba — Agent / operator guide

Use this file as the index for deployment and admin setup.

## Documentation

| Doc | Purpose |
|-----|---------|
| [VPS_DEPLOY.md](./VPS_DEPLOY.md) | **Start here** — `.env` variables, copy-paste VPS commands, nginx, admin seed |
| [WEB_DEPLOY.md](./WEB_DEPLOY.md) | Web dashboard, Telegram login, local dev proxy |
| [README.md](./README.md) | Project overview |

## Quick VPS commands

```bash
ssh sammy@kkooapp.co.tz
cd /home/sammy/marketmamba
cp .env.example .env && nano .env          # or scp .env from Mac (see VPS_DEPLOY.md)
docker compose -p marketmamba up -d --build
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/004_web_admins.sql
docker compose -p marketmamba exec app ./seedadmin
docker compose -p marketmamba logs -f app
```

## Admin login (email)

1. On the VPS `.env` set `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_TELEGRAM_ID` (your Telegram user id).
2. Run `seed-admin` (see above).
3. Open the site → **Admin login (email)**.

**Do not commit real passwords to git.** Set them only in `.env` on the server.

## Telegram admin

Same Telegram ID must be in `TELEGRAM_ADMIN_USER_IDS`. Commands: `/admin stats`, `/admin activate <id> <days>`.
