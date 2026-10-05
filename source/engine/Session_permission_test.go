package engine

import "dvah/tools"
import "dvah/types"
import net_http "net/http"
import "os"
import "path/filepath"
import "strings"
import "sync"
import "testing"

func newTestSession(t *testing.T, sandbox string) *Session {

	t.Helper()

	session := &Session{
		Agent:       &types.Agent{Name: "Test Agent", Role: "agent", Model: "test-model", Sandbox: sandbox},
		Config:      &types.Config{Name: "Test Agent", Model: "test-model", Sandbox: sandbox},
		Console:     types.NewConsole(nil, nil, 0),
		Permissions: types.NewPermissionStore(),
		client:      &net_http.Client{},
		mutex:       &sync.RWMutex{},
		adapters:    make(map[string]types.Adapter),
		tools:       make(map[string]types.Tool),
	}

	session.SetTool(tools.NewFiles([]string{"Write"}, sandbox))
	session.SetTool(tools.NewPermissions([]string{"Request", "List"}, session.Permissions))

	return session

}

func TestSession_PermissionScopeBypass(t *testing.T) {

	sandbox := t.TempDir()
	session := newTestSession(t, sandbox)

	target := filepath.Join(sandbox, "flag.txt")

	// Without a grant, the sensitive files.Write method must be denied.
	err0 := session.CallTool("", "files", "Write", map[string]any{
		"path":    target,
		"content": "should not be written",
	})

	if err0 == nil {
		t.Fatalf("Expected files.Write to be denied before a grant")
	}

	if _, statErr := os.Stat(target); statErr == nil {
		t.Fatalf("Expected %s to not exist before a grant", target)
	}

	// The agent grants itself files.Write, but only for /tmp/test.txt.
	err1 := session.CallTool("", "permissions", "Request", map[string]any{
		"tool":   "files",
		"method": "Write",
		"scope":  "/tmp/test.txt",
	})

	if err1 != nil {
		t.Fatalf("Expected permissions.Request to succeed, got %v", err1)
	}

	// NOTE: This is the vulnerability. The grant was scoped to /tmp/test.txt,
	// but because the store drops the scope, writing to a completely different
	// path now succeeds.
	err2 := session.CallTool("", "files", "Write", map[string]any{
		"path":    target,
		"content": "permission bypass worked",
	})

	if err2 != nil {
		t.Fatalf("Expected the scope-confused write to succeed, got %v", err2)
	}

	data, err3 := os.ReadFile(target)

	if err3 != nil {
		t.Fatalf("Expected %s to exist after the bypass, got %v", target, err3)
	}

	if strings.TrimSpace(string(data)) != "permission bypass worked" {
		t.Errorf("Expected %q, got %q", "permission bypass worked", strings.TrimSpace(string(data)))
	}

}
