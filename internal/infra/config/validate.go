package config

import (
	"fmt"
	"slices"
)

func Validate(cfg Config) error {
	if cfg.App.Port <= 0 {
		return fmt.Errorf("app.port must be > 0")
	}
	if !slices.Contains([]string{"sql", "integrated"}, cfg.Database.Mode) {
		return fmt.Errorf("database.mode must be sql or integrated")
	}
	if cfg.Database.Host == "" || cfg.Database.Name == "" {
		return fmt.Errorf("database.host and database.name are required")
	}
	if cfg.Database.Mode == "sql" && (cfg.Database.Username == "" || cfg.Database.Password == "") {
		return fmt.Errorf("database.username and database.password are required in sql mode")
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
