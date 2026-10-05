package types

import "dvah/schemas"
import utils_api_llamacpp "dvah/utils/api/llamacpp"
import utils_api_ollama "dvah/utils/api/ollama"
import utils_api_vllm "dvah/utils/api/vllm"
import utils_fmt "dvah/utils/fmt"
import "dvah/parsers/yaml"
import "encoding/json"
import "fmt"
import "io"
import net_url "net/url"
import "net/http"
import "os"
import "strings"

// Config carries the runtime settings for the harness. Agent identity (name,
// model, temperature) always comes from the agent.yml, while providers, tokens,
// endpoint URL, debug flag and the initial user prompt can be supplied by the
// global config (~/.config/dvah/config.yaml) and/or an optional per-invocation
// config.yml passed as the second CLI argument.
type Config struct {
	Name        string              `json:"name" yaml:"name"`
	Model       string              `json:"model" yaml:"model"`
	Prompt      string              `json:"prompt" yaml:"prompt"`
	Temperature float64             `json:"temperature" yaml:"temperature"`
	Sandbox     string              `json:"sandbox" yaml:"sandbox"`
	URL         *net_url.URL        `json:"url" yaml:"url"`
	Debug       bool                `json:"debug" yaml:"debug"`
	Providers   map[string]Provider `json:"providers" yaml:"providers"`
}

func NewConfig(name string, model string, prompt string, temperature float64, sandbox string, url *net_url.URL, debug bool) *Config {

	name   = strings.TrimSpace(name)
	model  = strings.TrimSpace(model)
	prompt = utils_fmt.FormatMultiLine(prompt)

	if temperature < 0.0 {
		temperature = 0.0
	} else if temperature > 1.0 {
		temperature = 1.0
	}

	if sandbox == "" {

		cwd, err := os.Getwd()

		if err == nil {
			sandbox = cwd
		} else {
			sandbox = "/tmp/dvah/sandbox"
		}

	}

	if url == nil {

		tmp, err := net_url.Parse("http://localhost:11434/v1")

		if err == nil {
			url = tmp
		}

	}

	return &Config{
		Name:        name,
		Model:       model,
		Prompt:      prompt,
		Temperature: temperature,
		Sandbox:     sandbox,
		URL:         url,
		Debug:       debug,
		Providers:   make(map[string]Provider),
	}

}

// LoadConfig reads a per-invocation config.yml from disk and parses it.
func LoadConfig(path string) (*Config, error) {

	resolved := strings.TrimSpace(path)

	if resolved == "" {
		return nil, fmt.Errorf("LoadConfig: Invalid config path")
	}

	data, err0 := os.ReadFile(resolved)

	if err0 != nil {
		return nil, fmt.Errorf("LoadConfig: %s", err0.Error())
	}

	config, err1 := ParseConfig(data)

	if err1 != nil {
		return nil, fmt.Errorf("LoadConfig: %s", err1.Error())
	}

	return config, nil

}

// Merge overlays runtime settings from another Config. Agent identity fields
// (Name, Model, Temperature) belong to the agent.yml and are never overwritten.
// URL, Debug, Prompt and Providers are taken from the overlay, so a
// per-invocation config.yml wins over the global config which wins over the
// built-in defaults.
func (config *Config) Merge(overlay *Config) {

	if overlay == nil {
		return
	}

	if overlay.URL != nil {
		config.URL = overlay.URL
	}

	if overlay.Debug == true {
		config.Debug = true
	}

	if strings.TrimSpace(overlay.Prompt) != "" {
		config.Prompt = overlay.Prompt
	}

	if len(overlay.Providers) > 0 {

		if config.Providers == nil {
			config.Providers = make(map[string]Provider)
		}

		for model, provider := range overlay.Providers {
			config.Providers[model] = provider
		}

	}

}

func ParseConfig(data []byte) (*Config, error) {

	if len(data) > 2 && data[0] == '{' && data[len(data)-1] == '}' {

		config := Config{}
		err    := json.Unmarshal(data, &config)

		if err == nil {
			return &config, nil
		} else {
			return nil, err
		}

	} else {

		config := Config{}
		err    := yaml.Unmarshal(data, &config)

		if err == nil {
			return &config, nil
		} else {
			return nil, err
		}

	}

}

func (config *Config) GetContextLength(model string) int {

	client       := &http.Client{}
	resolved_url := config.ResolveURL(model, "/models")

	request, err1 := http.NewRequest(http.MethodGet, resolved_url.String(), nil)

	if err1 == nil {

		response, err2 := client.Do(request)

		if err2 == nil {

			response_payload, err3 := io.ReadAll(response.Body)

			if err3 == nil {

				schema := schemas.ModelsResponse{}
				err4   := json.Unmarshal(response_payload, &schema)

				if err4 == nil {

					server_type := schema.OwnedBy()

					if server_type == "llamacpp" {

						return utils_api_llamacpp.GetContextLength(config.URL, config.Model)

					} else if server_type == "ollama" {

						return utils_api_ollama.GetContextLength(config.URL, config.Model)

					} else if server_type == "vllm" {

						return utils_api_vllm.GetContextLength(config.URL, config.Model)

					}

				}

			}

		}

	}

	return 0

}

func (config *Config) GetPrompt() string {
	return strings.TrimSpace(config.Prompt)
}

func (config *Config) HasProvider(model string) bool {

	_, ok := config.Providers[model]

	if ok == true {
		return true
	}

	return false

}

func (config *Config) ResolveToken(model string) string {

	provider, ok := config.Providers[model]

	if ok == true {
		return strings.TrimSpace(provider.Token)
	} else {
		return ""
	}

}

func (config *Config) ResolveModel(model string) string {

	provider, ok := config.Providers[model]

	if ok == true {

		if provider.Alias != "" {
			return provider.Alias
		} else {
			return model
		}

	} else {
		return model
	}

}

func (config *Config) ResolvePricing(model string) (Pricing, bool) {

	provider, ok := config.Providers[model]

	if ok == true && provider.Pricing != (Pricing{}) {
		return provider.Pricing, true
	}

	return Pricing{}, false

}

func (config *Config) ResolveURL(model string, path string) *net_url.URL {

	base_url := config.URL
	api_path := ""

	provider, ok := config.Providers[model]

	if ok == true && provider.URL != nil {
		base_url, _ = net_url.Parse(provider.URL.String())
	}

	if base_url == nil {
		base_url, _ = net_url.Parse("http://localhost:11434/v1")
	}

	if strings.HasPrefix(base_url.Path, "/") && len(base_url.Path) > 1 {

		// "/v1" or "/v1/"
		tmp_base := base_url.Path

		if strings.HasSuffix(tmp_base, "/") {
			tmp_base = strings.TrimSpace(tmp_base[0:len(tmp_base)-1])
		}

		// "/chat/completions"
		tmp_path := path

		if strings.HasPrefix(tmp_path, "/") {
			tmp_path = strings.TrimSpace(tmp_path[1:])
		}

		// "/v1/chat/completions"
		api_path = fmt.Sprintf("%s/%s", tmp_base, tmp_path)

	} else if strings.HasPrefix(path, "/") {
		api_path = strings.TrimSpace(path)
	}

	if api_path != "" {

		return base_url.ResolveReference(&net_url.URL{
			Path: api_path,
		})

	} else {
		return base_url
	}

}

func (config *Config) Public() *Config {

	clone := *config
	clone.Providers = make(map[string]Provider)

	return &clone

}

func (config Config) MarshalJSON() ([]byte, error) {

	url_str := ""

	if config.URL != nil {
		url_str = config.URL.String()
	}

	return json.Marshal(struct {
		Name        string              `json:"name"`
		Model       string              `json:"model"`
		Prompt      string              `json:"prompt"`
		Temperature float64             `json:"temperature"`
		Sandbox     string              `json:"sandbox"`
		URL         string              `json:"url"`
		Debug       bool                `json:"debug"`
		Providers   map[string]Provider `json:"providers,omitempty"`
	}{
		Name:        config.Name,
		Model:       config.Model,
		Prompt:      config.Prompt,
		Temperature: config.Temperature,
		Sandbox:     config.Sandbox,
		URL:         url_str,
		Debug:       config.Debug,
		Providers:   config.Providers,
	})

}

func (config *Config) UnmarshalJSON(data []byte) error {

	var tmp struct {
		Name        string              `json:"name"`
		Model       string              `json:"model"`
		Prompt      string              `json:"prompt"`
		Temperature float64             `json:"temperature"`
		Sandbox     string              `json:"sandbox"`
		URL         string              `json:"url"`
		Debug       bool                `json:"debug"`
		Providers   map[string]Provider `json:"providers"`
	}

	err0 := json.Unmarshal(data, &tmp)

	if err0 == nil {

		config.Name        = tmp.Name
		config.Model       = tmp.Model
		config.Prompt      = tmp.Prompt
		config.Temperature = tmp.Temperature
		config.Sandbox     = tmp.Sandbox
		config.Debug       = tmp.Debug
		config.Providers   = tmp.Providers

		tmp_url, err1 := net_url.Parse(tmp.URL)

		if err1 == nil {
			config.URL = tmp_url
		}

		return nil

	} else {
		return err0
	}

}

func (config *Config) MarshalYAML() ([]byte, error) {

	url_str := ""

	if config.URL != nil {
		url_str = config.URL.String()
	}

	return yaml.Marshal(struct {
		Name        string              `yaml:"name"`
		Model       string              `yaml:"model"`
		Prompt      string              `yaml:"prompt"`
		Temperature float64             `yaml:"temperature"`
		Sandbox     string              `yaml:"sandbox"`
		URL         string              `yaml:"url"`
		Debug       bool                `yaml:"debug"`
		Providers   map[string]Provider `yaml:"providers"`
	}{
		Name:        config.Name,
		Model:       config.Model,
		Prompt:      config.Prompt,
		Temperature: config.Temperature,
		Sandbox:     config.Sandbox,
		URL:         url_str,
		Debug:       config.Debug,
		Providers:   config.Providers,
	})

}

func (config *Config) UnmarshalYAML(data []byte) error {

	var tmp struct {
		Name        string              `yaml:"name"`
		Model       string              `yaml:"model"`
		Prompt      string              `yaml:"prompt"`
		Temperature float64             `yaml:"temperature"`
		Sandbox     string              `yaml:"sandbox"`
		URL         string              `yaml:"url"`
		Debug       bool                `yaml:"debug"`
		Providers   map[string]Provider `yaml:"providers"`
	}

	err0 := yaml.Unmarshal(data, &tmp)

	if err0 == nil {

		config.Name        = tmp.Name
		config.Model       = tmp.Model
		config.Prompt      = tmp.Prompt
		config.Temperature = tmp.Temperature
		config.Sandbox     = tmp.Sandbox
		config.Debug       = tmp.Debug
		config.Providers   = tmp.Providers

		tmp_url, err1 := net_url.Parse(tmp.URL)

		if err1 == nil {
			config.URL = tmp_url
		}

		return nil

	} else {
		return err0
	}

}
