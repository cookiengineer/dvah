package agents

import "os"
import "path/filepath"
import "strings"
import "testing"

func TestLoadAgent(t *testing.T) {

	dir, _ := os.MkdirTemp("/tmp", "dvah-test-agents-*")
	file    := filepath.Join(dir, "architect.yaml")

	yaml := `
name: "Peanut Architect"
role: architect
description: defines software specifications
model: some-abliterated-model
temperature: 0.5
allowed-programs:
  - go
allowed-tools:
  - files.Read
  - skills.Load
prompt: |
  Your name is "{{name}}" and your role is "{{role}}".
`

	err0 := os.WriteFile(file, []byte(yaml), 0666)

	if err0 != nil {
		t.Fatalf("Expected %v to be nil", err0)
	}

	agent, err1 := LoadAgent(file)

	if err1 != nil {
		t.Fatalf("Expected %v to be nil", err1)
	}

	if agent.Name != "Peanut Architect" {
		t.Errorf("Expected name %q, got %q", "Peanut Architect", agent.Name)
	}

	if agent.Role != "architect" {
		t.Errorf("Expected role %q, got %q", "architect", agent.Role)
	}

	if strings.Contains(agent.Prompt, "{{") == true {
		t.Errorf("Expected prompt template to be rendered, got %q", agent.Prompt)
	}

	if len(agent.Messages) != 1 || agent.Messages[0].Role != "system" {
		t.Errorf("Expected a single system message")
	}

	if len(agent.AllowedTools) != 2 {
		t.Errorf("Expected 2 allowed tools, got %d", len(agent.AllowedTools))
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

}

func TestLoadAgent_MissingFile(t *testing.T) {

	_, err := LoadAgent("/tmp/dvah-does-not-exist.yml")

	if err == nil {
		t.Errorf("Expected an error for a missing agent file")
	}

}
