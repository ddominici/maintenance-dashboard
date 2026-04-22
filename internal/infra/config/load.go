package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(path string) (Config, error) {
	cfg := defaults()

	_ = loadDotEnv(".env")

	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	applyEnvOverrides(&cfg)

	// Auto-migrate legacy single-server config into the servers list.
	if len(cfg.Servers) == 0 && cfg.Database.Host != "" {
		cfg.Servers = []ServerConfig{{Name: "default", Database: cfg.Database}}
	}

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func defaults() Config {
	return Config{
		App: AppConfig{
			Name:                "maintenance-dashboard",
			Env:                 "development",
			Host:                "0.0.0.0",
			Port:                8080,
			ReadTimeoutSeconds:  15,
			WriteTimeoutSeconds: 15,
			IdleTimeoutSeconds:  60,
		},
		Auth: AuthConfig{Enabled: true, Username: "admin", Password: "change-me"},
		Database: DatabaseConfig{
			Mode:                     "sql",
			Host:                     "localhost",
			Port:                     1433,
			Name:                     "MaintenanceDB",
			Encrypt:                  false,
			TrustServerCertificate:   true,
			ConnectionTimeoutSeconds: 10,
			MaxOpenConns:             20,
			MaxIdleConns:             10,
			ConnMaxLifetimeMinutes:   30,
		},
		Cache: CacheConfig{Enabled: true, DefaultTTLSeconds: 60, DashboardTTLSeconds: 30, DetailTTLSeconds: 60, FiltersTTLSeconds: 300, CleanupIntervalSeconds: 60},
		UI:    UIConfig{DefaultLanguage: "en", SupportedLanguages: []string{"en", "it"}},
	}
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return s.Err()
}

func applyEnvOverrides(cfg *Config) {
	setString := func(env string, target *string) {
		if v := os.Getenv(env); v != "" {
			*target = v
		}
	}
	setInt := func(env string, target *int) {
		if v := os.Getenv(env); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				*target = n
			}
		}
	}
	setBool := func(env string, target *bool) {
		if v := os.Getenv(env); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				*target = b
			}
		}
	}

	setString("APP_NAME", &cfg.App.Name)
	setString("APP_ENV", &cfg.App.Env)
	setString("APP_HOST", &cfg.App.Host)
	setInt("APP_PORT", &cfg.App.Port)

	setBool("AUTH_ENABLED", &cfg.Auth.Enabled)
	setString("AUTH_USERNAME", &cfg.Auth.Username)
	setString("AUTH_PASSWORD", &cfg.Auth.Password)

	setString("DATABASE_MODE", &cfg.Database.Mode)
	setString("DATABASE_HOST", &cfg.Database.Host)
	setInt("DATABASE_PORT", &cfg.Database.Port)
	setString("DATABASE_INSTANCE", &cfg.Database.Instance)
	setString("DATABASE_NAME", &cfg.Database.Name)
	setString("DATABASE_USERNAME", &cfg.Database.Username)
	setString("DATABASE_PASSWORD", &cfg.Database.Password)
	setBool("DATABASE_ENCRYPT", &cfg.Database.Encrypt)
	setBool("DATABASE_TRUST_SERVER_CERTIFICATE", &cfg.Database.TrustServerCertificate)
	setInt("DATABASE_CONNECTION_TIMEOUT_SECONDS", &cfg.Database.ConnectionTimeoutSeconds)
	setInt("DATABASE_MAX_OPEN_CONNS", &cfg.Database.MaxOpenConns)
	setInt("DATABASE_MAX_IDLE_CONNS", &cfg.Database.MaxIdleConns)
	setInt("DATABASE_CONN_MAX_LIFETIME_MINUTES", &cfg.Database.ConnMaxLifetimeMinutes)

	setBool("CACHE_ENABLED", &cfg.Cache.Enabled)
	setInt("CACHE_DEFAULT_TTL_SECONDS", &cfg.Cache.DefaultTTLSeconds)
	setInt("CACHE_DASHBOARD_TTL_SECONDS", &cfg.Cache.DashboardTTLSeconds)
	setInt("CACHE_DETAIL_TTL_SECONDS", &cfg.Cache.DetailTTLSeconds)
	setInt("CACHE_FILTERS_TTL_SECONDS", &cfg.Cache.FiltersTTLSeconds)
	setInt("CACHE_CLEANUP_INTERVAL_SECONDS", &cfg.Cache.CleanupIntervalSeconds)

	for i := range cfg.Servers {
		prefix := "SERVER_" + serverEnvName(cfg.Servers[i].Name) + "_"
		setString(prefix+"HOST", &cfg.Servers[i].Database.Host)
		setInt(prefix+"PORT", &cfg.Servers[i].Database.Port)
		setString(prefix+"INSTANCE", &cfg.Servers[i].Database.Instance)
		setString(prefix+"MODE", &cfg.Servers[i].Database.Mode)
		setString(prefix+"NAME", &cfg.Servers[i].Database.Name)
		setString(prefix+"USERNAME", &cfg.Servers[i].Database.Username)
		setString(prefix+"PASSWORD", &cfg.Servers[i].Database.Password)
		setBool(prefix+"ENCRYPT", &cfg.Servers[i].Database.Encrypt)
		setBool(prefix+"TRUST_SERVER_CERTIFICATE", &cfg.Servers[i].Database.TrustServerCertificate)
	}

	setString("UI_DEFAULT_LANGUAGE", &cfg.UI.DefaultLanguage)
	if v := os.Getenv("UI_SUPPORTED_LANGUAGES"); v != "" {
		parts := strings.Split(v, ",")
		langs := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				langs = append(langs, p)
			}
		}
		if len(langs) > 0 {
			cfg.UI.SupportedLanguages = langs
		}
	}
}

// serverEnvName normalizes a server name for use in env var names.
// "SQL 2019-prod" → "SQL_2019_PROD"
func serverEnvName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(name) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
