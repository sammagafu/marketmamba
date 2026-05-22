/** Mirrors internal/auth/acl.go permission names. */
export const Perm = {
  statusView: 'status:view',
  accountView: 'account:view',
  positionsView: 'positions:view',
  tradesView: 'trades:view',
  subscriptionView: 'subscription:view',
  brokerView: 'broker:view',
  brokerManage: 'broker:manage',
  brokerTest: 'broker:test',
  adminStats: 'admin:stats',
  adminUsersView: 'admin:users:view',
  adminUsersBlock: 'admin:users:block',
  adminUsersRevoke: 'admin:users:revoke',
  adminActivate: 'admin:activate',
  adminTradesView: 'admin:trades:view',
  adminSignalsBroadcast: 'admin:signals:broadcast',
}

export function can(permissions, perm) {
  return Array.isArray(permissions) && permissions.includes(perm)
}

export function isAdminRole(role) {
  return role === 'admin'
}

export function applyProfile(state, profile) {
  state.role = profile.role || 'user'
  state.isAdmin = !!profile.is_admin
  state.permissions = profile.permissions || []
  state.isBlocked = !!profile.is_blocked
  state.canTrade = profile.can_trade !== false
  state.tradeMessage = profile.trade_message || ''
}
