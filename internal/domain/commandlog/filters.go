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

// SortSpec selects which column to order a result set by, and in which
// direction. By is a logical key (e.g. "lastFragmentation") that each
// repository maps to a whitelisted SQL column; an empty or unrecognized By
// falls back to the query's default ordering. Dir is "asc" or "desc"
// (default "desc"). Sorting is applied server-side over the full dataset
// before the TOP/limit, so the limit captures the true top rows.
type SortSpec struct {
	By  string `json:"by,omitempty"`
	Dir string `json:"dir,omitempty"`
}
