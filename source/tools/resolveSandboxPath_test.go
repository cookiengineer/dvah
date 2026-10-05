package tools

import "testing"

func TestResolveSandboxPath_RootSandbox(t *testing.T) {

	resolved1, err1 := resolveSandboxPath("/", "/tmp/test.txt")

	if err1 != nil {
		t.Fatalf("Expected %v to be nil", err1)
	}

	if resolved1 != "/tmp/test.txt" {
		t.Errorf("Expected %q, got %q", "/tmp/test.txt", resolved1)
	}

	resolved2, err2 := resolveSandboxPath("/", "/etc/flag.txt")

	if err2 != nil {
		t.Fatalf("Expected %v to be nil", err2)
	}

	if resolved2 != "/etc/flag.txt" {
		t.Errorf("Expected %q, got %q", "/etc/flag.txt", resolved2)
	}

}

func TestResolveSandboxPath_Escape(t *testing.T) {

	_, err := resolveSandboxPath("/tmp", "/etc/flag.txt")

	if err == nil {
		t.Errorf("Expected an error when escaping the sandbox")
	}

}
