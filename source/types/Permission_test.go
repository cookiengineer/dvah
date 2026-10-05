package types

import "testing"

func TestPermissionStore_ScopeIsIgnored(t *testing.T) {

	store := NewPermissionStore()

	if store.Required("files", "Write") != true {
		t.Errorf("Expected files.Write to require a grant")
	}

	if store.Granted("files", "Write") != false {
		t.Errorf("Expected files.Write to be denied before any grant")
	}

	store.Request("files", "Write", "/tmp/test.txt")

	if store.Granted("files", "Write") != true {
		t.Errorf("Expected files.Write to be granted after the request")
	}

	// NOTE: This is the vulnerability. The grant was requested for the scope
	// /tmp/test.txt, but Granted() has no scope argument and the store discards
	// it, so the method is now authorised for every path.
	if len(store.Grants()) != 1 {
		t.Errorf("Expected exactly one grant")
	}

	if store.Granted("files", "Read") != false {
		t.Errorf("Expected unrelated methods to stay denied")
	}

}

func TestPermissionStore_Revoke(t *testing.T) {

	store := NewPermissionStore()

	store.Request("files", "Write", "/tmp/test.txt")
	store.Revoke("files", "Write")

	if store.Granted("files", "Write") != false {
		t.Errorf("Expected grant to be revoked")
	}

}
