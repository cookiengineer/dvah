package types

import net_url "net/url"
import "os"
import "path/filepath"
import "strings"
import "testing"

func TestConfig_ResolvePricing_FromProvider(t *testing.T) {

	config := &Config{
		Providers: map[string]Provider{
			"deepseek-v4-pro:cloud": {
				Model:  "deepseek-v4-pro:cloud",
				Alias:  "deepseek-v4-pro",
				Pricing: Pricing{
					InputPrice:       1.32,
					OutputPrice:      3.96,
					CachedInputPrice: 0.044,
				},
			},
		},
	}

	pricing, ok := config.ResolvePricing("deepseek-v4-pro:cloud")

	if ok != true {
		t.Errorf("Expected pricing to be resolved")
	}

	if pricing.InputPrice != 1.32 {
		t.Errorf("Expected InputPrice %v to be %v", pricing.InputPrice, 1.32)
	}

	if pricing.OutputPrice != 3.96 {
		t.Errorf("Expected OutputPrice %v to be %v", pricing.OutputPrice, 3.96)
	}

	if pricing.CachedInputPrice != 0.044 {
		t.Errorf("Expected CachedInputPrice %v to be %v", pricing.CachedInputPrice, 0.044)
	}

}

func TestConfig_ResolvePricing_Unknown(t *testing.T) {

	config := &Config{
		Providers: map[string]Provider{},
	}

	pricing, ok := config.ResolvePricing("some-model")

	if ok == true {
		t.Errorf("Expected pricing to be unresolved, got %v", pricing)
	}

}

func TestParseProvider_YAML(t *testing.T) {

	data := []byte(`
model: deepseek-v4-pro:cloud
url: https://api.deepseek.com
alias: deepseek-v4-pro
pricing:
  input_price: 1.32
  output_price: 3.96
  cached_input_price: 0.044
`)

	provider, err := ParseProvider(data)

	if err != nil {
		t.Errorf("Expected %v to be nil", err)
	}

	if provider == nil {
		t.Fatalf("Expected provider to be not nil")
	}

	if provider.Model != "deepseek-v4-pro:cloud" {
		t.Errorf("Expected Model %q to be %q", provider.Model, "deepseek-v4-pro:cloud")
	}

	if provider.Alias != "deepseek-v4-pro" {
		t.Errorf("Expected Alias %q to be %q", provider.Alias, "deepseek-v4-pro")
	}

	if provider.URL == nil {
		t.Errorf("Expected URL to be not nil")
	} else if provider.URL.String() != "https://api.deepseek.com" {
		t.Errorf("Expected URL %q to be %q", provider.URL.String(), "https://api.deepseek.com")
	}

	if provider.Pricing.InputPrice != 1.32 {
		t.Errorf("Expected InputPrice %v to be %v", provider.Pricing.InputPrice, 1.32)
	}

	if provider.Pricing.OutputPrice != 3.96 {
		t.Errorf("Expected OutputPrice %v to be %v", provider.Pricing.OutputPrice, 3.96)
	}

	if provider.Pricing.CachedInputPrice != 0.044 {
		t.Errorf("Expected CachedInputPrice %v to be %v", provider.Pricing.CachedInputPrice, 0.044)
	}

}

func TestParseProvider_JSON(t *testing.T) {

	// NOTE: JSON does not deserialize pricing by design (see Provider.MarshalJSON)
	data := []byte(`{"model":"deepseek-v4-pro:cloud","url":"https://api.deepseek.com","alias":"deepseek-v4-pro","token":"sk-test"}`)

	provider, err := ParseProvider(data)

	if err != nil {
		t.Errorf("Expected %v to be nil", err)
	}

	if provider == nil {
		t.Fatalf("Expected provider to be not nil")
	}

	if provider.Model != "deepseek-v4-pro:cloud" {
		t.Errorf("Expected Model %q to be %q", provider.Model, "deepseek-v4-pro:cloud")
	}

	if provider.Alias != "deepseek-v4-pro" {
		t.Errorf("Expected Alias %q to be %q", provider.Alias, "deepseek-v4-pro")
	}

	if provider.Token != "sk-test" {
		t.Errorf("Expected Token %q to be %q", provider.Token, "sk-test")
	}

	if provider.URL == nil {
		t.Errorf("Expected URL to be not nil")
	}

	if provider.Pricing != (Pricing{}) {
		t.Errorf("Expected Pricing %v to be empty", provider.Pricing)
	}

}

func TestProvider_MarshalJSON_OmitsPricing(t *testing.T) {

	url, _ := net_url.Parse("https://api.deepseek.com")

	provider := Provider{
		Model: "deepseek-v4-pro:cloud",
		URL:   url,
		Alias: "deepseek-v4-pro",
		Pricing: Pricing{
			InputPrice:  1.32,
			OutputPrice: 3.96,
		},
	}

	data, err := provider.MarshalJSON()

	if err != nil {
		t.Errorf("Expected %v to be nil", err)
	}

	parsed, err := ParseProvider(data)

	if err != nil {
		t.Errorf("Expected %v to be nil", err)
	}

	if parsed == nil {
		t.Fatalf("Expected provider to be not nil")
	}

	if parsed.Pricing != (Pricing{}) {
		t.Errorf("Expected Pricing %v to be empty after JSON round-trip", parsed.Pricing)
	}

	if parsed.Model != "deepseek-v4-pro:cloud" {
		t.Errorf("Expected Model %q to be %q", parsed.Model, "deepseek-v4-pro:cloud")
	}

}

func TestConfig_Merge(t *testing.T) {

	base_url, _ := net_url.Parse("http://localhost:11434/v1")
	over_url, _ := net_url.Parse("https://api.deepseek.com")

	base := &Config{
		Name:      "Agent Name",
		Model:     "agent-model",
		Providers: make(map[string]Provider),
	}

	overlay := &Config{
		Name:      "Should Not Override",
		Model:     "should-not-override",
		URL:       over_url,
		Debug:     true,
		Prompt:    "first user prompt",
		Providers: map[string]Provider{
			"agent-model": {
				Model: "agent-model",
				URL:   base_url,
				Token: "sk-test",
			},
		},
	}

	base.Merge(overlay)

	if base.Name != "Agent Name" {
		t.Errorf("Expected Name to stay %q, got %q", "Agent Name", base.Name)
	}

	if base.Model != "agent-model" {
		t.Errorf("Expected Model to stay %q, got %q", "agent-model", base.Model)
	}

	if base.URL != over_url {
		t.Errorf("Expected URL to be overridden")
	}

	if base.Debug != true {
		t.Errorf("Expected Debug to be true")
	}

	if base.GetPrompt() != "first user prompt" {
		t.Errorf("Expected prompt %q, got %q", "first user prompt", base.GetPrompt())
	}

	if base.ResolveToken("agent-model") != "sk-test" {
		t.Errorf("Expected provider token to be merged")
	}

}

func TestConfig_LoadConfig(t *testing.T) {

	dir, _ := os.MkdirTemp("/tmp", "dvah-test-config-*")
	file    := filepath.Join(dir, "config.yml")

	data := `
url: https://api.deepseek.com
debug: true
prompt: |
  First user prompt line one
  First user prompt line two
providers:
  deepseek-v4-pro:cloud:
    model: deepseek-v4-pro:cloud
    url: https://api.deepseek.com
    alias: deepseek-chat
    token: "sk-abc"
`

	err0 := os.WriteFile(file, []byte(data), 0666)

	if err0 != nil {
		t.Fatalf("Expected %v to be nil", err0)
	}

	config, err1 := LoadConfig(file)

	if err1 != nil {
		t.Fatalf("Expected %v to be nil", err1)
	}

	if config.URL == nil || config.URL.Host != "api.deepseek.com" {
		t.Errorf("Expected URL host %q, got %v", "api.deepseek.com", config.URL)
	}

	if config.Debug != true {
		t.Errorf("Expected Debug to be true")
	}

	if strings.Contains(config.Prompt, "line one") == false || strings.Contains(config.Prompt, "line two") == false {
		t.Errorf("Expected multi-line prompt to be preserved, got %q", config.Prompt)
	}

	if config.ResolveToken("deepseek-v4-pro:cloud") != "sk-abc" {
		t.Errorf("Expected provider token to be parsed")
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

}

