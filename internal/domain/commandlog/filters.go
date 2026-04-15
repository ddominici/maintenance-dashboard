package commandlog

import "time"

type QueryFilters struct {
	DateFrom    *time.Time `json:"dateFrom,omitempty"`
	DateTo      *time.Time `json:"dateTo,omitempty"`
	Database    string     `json:"database,omitempty"`
	Schema      string     `json:"schema,omitempty"`
	Object      string     `json:"object,omitempty"`
	CommandType string     `json:"commandType,omitempty"`
	OnlyErrors  bool       `json:"onlyErrors,omitempty"`
}
