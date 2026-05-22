package auth

// Role identifies the user's access tier.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// Permission names used in API responses and (future) policy checks.
const (
	PermStatusView       = "status:view"
	PermAccountView      = "account:view"
	PermPositionsView    = "positions:view"
	PermTradesView       = "trades:view"
	PermSubscriptionView = "subscription:view"
	PermBrokerView       = "broker:view"
	PermBrokerManage     = "broker:manage"
	PermBrokerTest       = "broker:test"

	PermAdminStats           = "admin:stats"
	PermAdminUsersView       = "admin:users:view"
	PermAdminUsersBlock      = "admin:users:block"
	PermAdminUsersRevoke     = "admin:users:revoke"
	PermAdminActivate        = "admin:activate"
	PermAdminTradesView      = "admin:trades:view"
	PermAdminSignalsBroadcast = "admin:signals:broadcast"
)

var userPermissions = []string{
	PermStatusView,
	PermAccountView,
	PermPositionsView,
	PermTradesView,
	PermSubscriptionView,
	PermBrokerView,
	PermBrokerManage,
	PermBrokerTest,
}

var adminPermissions = append(append([]string{}, userPermissions...),
	PermAdminStats,
	PermAdminUsersView,
	PermAdminUsersBlock,
	PermAdminUsersRevoke,
	PermAdminActivate,
	PermAdminTradesView,
	PermAdminSignalsBroadcast,
)

// ResolveRole returns admin or user for a Telegram ID.
func ResolveRole(isAdmin bool) Role {
	if isAdmin {
		return RoleAdmin
	}
	return RoleUser
}

// PermissionsFor returns the permission strings for a role.
func PermissionsFor(role Role) []string {
	if role == RoleAdmin {
		out := make([]string, len(adminPermissions))
		copy(out, adminPermissions)
		return out
	}
	out := make([]string, len(userPermissions))
	copy(out, userPermissions)
	return out
}

// HasPermission checks if role grants a permission.
func HasPermission(role Role, perm string) bool {
	for _, p := range PermissionsFor(role) {
		if p == perm {
			return true
		}
	}
	return false
}

// Profile is returned by /auth/me for clients to enforce UI ACL.
type Profile struct {
	TelegramID   int64    `json:"telegram_id"`
	Role         Role     `json:"role"`
	IsAdmin      bool     `json:"is_admin"`
	Permissions  []string `json:"permissions"`
	IsBlocked    bool     `json:"is_blocked"`
	CanTrade     bool     `json:"can_trade"`
	TradeMessage string   `json:"trade_message,omitempty"`
}
