package db

import (
	"fmt"
	"net/url"
	"strconv"

	"maintenance-dashboard/internal/infra/config"
)

func BuildDSN(cfg config.DatabaseConfig) string {
	q := url.Values{}
	q.Set("database", cfg.Name)
	if cfg.Encrypt {
		q.Set("encrypt", "true")
	} else {
		q.Set("encrypt", "disable")
	}
	if cfg.TrustServerCertificate {
		q.Set("TrustServerCertificate", "true")
	}
	if cfg.ConnectionTimeoutSeconds > 0 {
		q.Set("connection timeout", strconv.Itoa(cfg.ConnectionTimeoutSeconds))
	}

	host := cfg.Host
	instance := ""
	if cfg.Instance != "" {
		instance = cfg.Instance
	} else if cfg.Port > 0 {
		host = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}

	path := ""
	if instance != "" {
		path = "/" + instance
	}

	if cfg.Mode == "integrated" {
		q.Set("trusted_connection", "yes")
		return fmt.Sprintf("sqlserver://@%s%s?%s", host, path, q.Encode())
	}

	userInfo := url.UserPassword(cfg.Username, cfg.Password)
	return fmt.Sprintf("sqlserver://%s@%s%s?%s", userInfo, host, path, q.Encode())
}
