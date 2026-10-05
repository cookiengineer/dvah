package yaml

import "strings"
import "testing"

type commentDoc struct {
	Name   string            `yaml:"name"`
	Prompt string            `yaml:"prompt"`
	List   []string          `yaml:"list"`
	Nested map[string]string `yaml:"nested"`
}

func TestComments_FullLineAndTrailing(t *testing.T) {

	input := `
# leading comment
name: peanut # trailing comment
prompt: |
  line one
  # this hash is part of the literal block scalar
list:
  - go # trailing item comment
  - gofmt
nested:
  # nested comment
  key: value # trailing value comment
`

	doc := commentDoc{}
	err := Unmarshal([]byte(input), &doc)

	if err != nil {
		t.Fatalf("Expected %v to be nil", err)
	}

	if doc.Name != "peanut" {
		t.Errorf("Expected name %q, got %q", "peanut", doc.Name)
	}

	if strings.Contains(doc.Prompt, "line one") == false {
		t.Errorf("Expected prompt to contain %q, got %q", "line one", doc.Prompt)
	}

	if strings.Contains(doc.Prompt, "# this hash is part of the literal block scalar") == false {
		t.Errorf("Expected hash inside block scalar to be preserved, got %q", doc.Prompt)
	}

	if len(doc.List) != 2 || doc.List[0] != "go" || doc.List[1] != "gofmt" {
		t.Errorf("Expected list [go gofmt], got %v", doc.List)
	}

	if doc.Nested["key"] != "value" {
		t.Errorf("Expected nested[key] %q, got %q", "value", doc.Nested["key"])
	}

}

func TestComments_ColonInCommentDoesNotStealFields(t *testing.T) {

	input := `
# note:
name: peanut
nested:
  # inner note:
  key: value
`

	doc := commentDoc{}
	err := Unmarshal([]byte(input), &doc)

	if err != nil {
		t.Fatalf("Expected %v to be nil", err)
	}

	if doc.Name != "peanut" {
		t.Errorf("Expected name %q, got %q", "peanut", doc.Name)
	}

	if doc.Nested["key"] != "value" {
		t.Errorf("Expected nested[key] %q, got %q", "value", doc.Nested["key"])
	}

}
