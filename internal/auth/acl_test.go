package auth

import "testing"

func TestHasPermission(t *testing.T) {
	if !HasPermission(RoleUser, PermTradesView) {
		t.Fatal("user should view trades")
	}
	if HasPermission(RoleUser, PermAdminStats) {
		t.Fatal("user must not have admin stats")
	}
	if !HasPermission(RoleAdmin, PermAdminStats) {
		t.Fatal("admin should have admin stats")
	}
	if !HasPermission(RoleAdmin, PermTradesView) {
		t.Fatal("admin inherits user perms")
	}
}
