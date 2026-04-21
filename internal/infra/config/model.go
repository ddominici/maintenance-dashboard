package config

type Config struct {
	App      AppConfig      `yaml:"app"`
	Auth     AuthConfig     `yaml:"auth"`
	Database DatabaseConfig `yaml:"database"` // legacy single-server; auto-migrated to Servers on load
	Servers  []ServerConfig `yaml:"servers"`
	Cache    CacheConfig    `yaml:"cache"`
	UI       UIConfig       `yaml:"ui"`
}

type ServerConfig struct {
	Name     string         `yaml:"name"`
	Database DatabaseConfig `yaml:"database"`
}

type AppConfig struct {
	Name                string `yaml:"name"`
	Env                 string `yaml:"env"`
	Host                string `yaml:"host"`
	Port                int    `yaml:"port"`
	ReadTimeoutSeconds  int    `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `yaml:"write_timeout_seconds"`
	IdleTimeoutSeconds  int    `yaml:"idle_timeout_seconds"`
}

type AuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type DatabaseConfig struct {
	Mode                     string `yaml:"mode"`
	Host                     string `yaml:"host"`
	Port                     int    `yaml:"port"`
	Instance                 string `yaml:"instance"`
	Name                     string `yaml:"name"`
	Username                 string `yaml:"username"`
	Password                 string `yaml:"password"`
	Encrypt                  bool   `yaml:"encrypt"`
	TrustServerCertificate   bool   `yaml:"trust_server_certificate"`
	ConnectionTimeoutSeconds int    `yaml:"connection_timeout_seconds"`
	MaxOpenConns             int    `yaml:"max_open_conns"`
	MaxIdleConns             int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeMinutes   int    `yaml:"conn_max_lifetime_minutes"`
}

type CacheConfig struct {
	Enabled                bool `yaml:"enabled"`
	DefaultTTLSeconds      int  `yaml:"default_ttl_seconds"`
	DashboardTTLSeconds    int  `yaml:"dashboard_ttl_seconds"`
	DetailTTLSeconds       int  `yaml:"detail_ttl_seconds"`
	FiltersTTLSeconds      int  `yaml:"filters_ttl_seconds"`
	CleanupIntervalSeconds int  `yaml:"cleanup_interval_seconds"`
}

type UIConfig struct {
	DefaultLanguage    string   `yaml:"default_language"`
	SupportedLanguages []string `yaml:"supported_languages"`
}
