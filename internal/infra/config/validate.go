package config

import (
	"fmt"
	"slices"
)

func Validate(cfg Config) error {
	if cfg.App.Port <= 0 {
		return fmt.Errorf("app.port must be > 0")
	}
	if len(cfg.Servers) == 0 {
		return fmt.Errorf("at least one server must be configured under 'servers' or 'database'")
	}
	names := map[string]bool{}
	for _, s := range cfg.Servers {
		if s.Name == "" {
			return fmt.Errorf("server name cannot be empty")
		}
		if names[s.Name] {
			return fmt.Errorf("duplicate server name: %q", s.Name)
		}
		names[s.Name] = true
		if !slices.Contains([]string{"sql", "integrated"}, s.Database.Mode) {
			return fmt.Errorf("server %q: database.mode must be sql or integrated", s.Name)
		}
		if s.Database.Host == "" || s.Database.Name == "" {
			return fmt.Errorf("server %q: database.host and database.name are required", s.Name)
		}
		if s.Database.Mode == "sql" && (s.Database.Username == "" || s.Database.Password == "") {
			return fmt.Errorf("server %q: database.username and database.password are required in sql mode", s.Name)
		}
	}
	if cfg.Auth.Enabled && (cfg.Auth.Username == "" || cfg.Auth.Password == "") {
		return fmt.Errorf("auth.username and auth.password are required when auth is enabled")
	}
	if len(cfg.UI.SupportedLanguages) == 0 {
		return fmt.Errorf("ui.supported_languages cannot be empty")
	}
	if !slices.Contains(cfg.UI.SupportedLanguages, cfg.UI.DefaultLanguage) {
		return fmt.Errorf("ui.default_language must be in supported_languages")
	}
	return nil
}
