package core

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ProviderName string
	Model        string
	BaseURL      string
	APIKey       string
	Prompt       string
}

func ConfigFromArgs(args []string, stdin io.Reader) (Config, error) {
	fs := flag.NewFlagSet("holy", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := loadSettingsConfig()
	flagProvider := cfg.ProviderName
	flagModel := cfg.Model
	flagBaseURL := cfg.BaseURL
	flagAPIKey := cfg.APIKey
	fs.StringVar(&flagProvider, "provider", "", "provider name")
	fs.StringVar(&flagModel, "model", "", "model name")
	fs.StringVar(&flagBaseURL, "base-url", "", "provider base URL")
	fs.StringVar(&flagAPIKey, "api-key", "", "provider API key")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	visited := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})
	if visited["provider"] {
		cfg.ProviderName = flagProvider
	}
	if visited["model"] {
		cfg.Model = flagModel
	}
	if visited["base-url"] {
		cfg.BaseURL = flagBaseURL
	}
	if visited["api-key"] {
		cfg.APIKey = flagAPIKey
	}

	prompt := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if prompt == "" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return Config{}, fmt.Errorf("read prompt from stdin: %w", err)
		}
		prompt = strings.TrimSpace(string(data))
	}

	if prompt == "" {
		return Config{}, fmt.Errorf("prompt is required via argv or stdin")
	}

	cfg.Prompt = prompt
	applyProviderDefaults(&cfg)
	if err := validateProviderConfig(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

type settingsConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
}

func loadSettingsConfig() Config {
	cfg := Config{}
	for _, path := range settingsPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var fileCfg settingsConfig
		if err := json.Unmarshal(data, &fileCfg); err != nil {
			continue
		}
		if fileCfg.Provider != "" {
			cfg.ProviderName = fileCfg.Provider
		}
		if fileCfg.Model != "" {
			cfg.Model = fileCfg.Model
		}
		if fileCfg.BaseURL != "" {
			cfg.BaseURL = fileCfg.BaseURL
		}
		if fileCfg.APIKey != "" {
			cfg.APIKey = fileCfg.APIKey
		}
	}
	return cfg
}

func settingsPaths() []string {
	paths := make([]string, 0, 2)
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		paths = append(paths, filepath.Join(home, ".holy", "settings.json"))
	}
	if cwd, err := os.Getwd(); err == nil && strings.TrimSpace(cwd) != "" {
		paths = append(paths, filepath.Join(cwd, ".holy", "settings.json"))
	}
	return paths
}

func applyProviderDefaults(cfg *Config) {
	if cfg == nil {
		return
	}
	switch cfg.ProviderName {
	case "anthropic":
		if cfg.APIKey == "" {
			cfg.APIKey = strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
		}
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://api.anthropic.com"
		}
	case "openai-responses":
		if cfg.APIKey == "" {
			cfg.APIKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
		}
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://api.openai.com/v1"
		}
	case "openai-compatible":
		if cfg.APIKey == "" {
			cfg.APIKey = strings.TrimSpace(os.Getenv("OPENAI_COMPATIBLE_API_KEY"))
		}
		if cfg.APIKey == "" {
			cfg.APIKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
		}
	}
}

func validateProviderConfig(cfg Config) error {
	switch cfg.ProviderName {
	case "anthropic":
		if strings.TrimSpace(cfg.Model) == "" {
			return fmt.Errorf("anthropic model is required")
		}
		if strings.TrimSpace(cfg.APIKey) == "" {
			return fmt.Errorf("anthropic api key is required")
		}
	case "openai-responses":
		if strings.TrimSpace(cfg.Model) == "" {
			return fmt.Errorf("openai-responses model is required")
		}
		if strings.TrimSpace(cfg.APIKey) == "" {
			return fmt.Errorf("openai-responses api key is required")
		}
	case "openai-compatible":
		if strings.TrimSpace(cfg.Model) == "" {
			return fmt.Errorf("openai-compatible model is required")
		}
		if strings.TrimSpace(cfg.APIKey) == "" {
			return fmt.Errorf("openai-compatible api key is required")
		}
		if strings.TrimSpace(cfg.BaseURL) == "" {
			return fmt.Errorf("openai-compatible base URL is required")
		}
	}
	return nil
}
