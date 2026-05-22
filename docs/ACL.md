# Access control (ACL)

## Roles

| Role | `role` field | Source |
|------|--------------|--------|
| Admin | `admin` | `TELEGRAM_ADMIN_USER_IDS` in `.env` |
| Trader | `user` | Default for all other logged-in users |

Email login is **admin-only**: account must exist in `web_admins` and the linked Telegram ID must be in `TELEGRAM_ADMIN_USER_IDS`.

## Permissions

Returned on login and `GET /api/v1/auth/me` as `permissions[]`.

**Trader**

- `status:view`, `account:view`, `positions:view`, `trades:view`
- `subscription:view`, `broker:view`, `broker:manage`, `broker:test`

**Admin** (includes all trader permissions plus)

- `admin:stats`, `admin:users:view`, `admin:users:block`, `admin:users:revoke`
- `admin:activate`, `admin:trades:view`, `admin:signals:broadcast`

## HTTP routes

| Path | Middleware |
|------|------------|
| `/api/v1/auth/*` (public login) | none |
| `/api/v1/admin/*` | `withAdmin` |
| All other `/api/v1/*` (authenticated) | `withUser` |

`withUser` rejects blocked accounts (except admins). `withAdmin` requires admin role.

## Blocked users

Set `is_blocked` via admin panel or storage. Blocked traders:

- Cannot call authenticated API (403)
- See blocked banner on web; trading UI hidden
- Telegram bot replies with blocked message

Admins are never treated as blocked.

## Frontend

`web/src/acl.js` — `can(permissions, perm)` mirrors server permission names.

Admin UI loads only when `can(permissions, 'admin:stats')`.

## Code

- `internal/auth/acl.go` — roles and permission lists
- `internal/api/acl_middleware.go` — `withUser`, `withAdmin`, `buildACLProfile`
