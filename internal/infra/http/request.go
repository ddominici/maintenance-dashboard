package httpx

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"maintenance-dashboard/internal/domain/commandlog"
)

func resolveServer[T any](services map[string]T, r *http.Request) (T, error) {
	name := r.URL.Query().Get("server")
	svc, ok := services[name]
	if !ok {
		var zero T
		return zero, fmt.Errorf("unknown server: %q", name)
	}
	return svc, nil
}

func ParseFilters(r *http.Request) (commandlog.QueryFilters, error) {
	q := r.URL.Query()
	var f commandlog.QueryFilters
	if v := q.Get("dateFrom"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return f, err
		}
		f.DateFrom = &t
	}
	if v := q.Get("dateTo"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return f, err
		}
		t = t.Add(24 * time.Hour)
		f.DateTo = &t
	}
	f.Database = q.Get("database")
	f.Schema = q.Get("schema")
	f.Object = q.Get("object")
	f.CommandType = q.Get("commandType")
	if v := q.Get("onlyErrors"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return f, err
		}
		f.OnlyErrors = b
	}
	return f, nil
}

// ParseSort reads the optional sortBy/sortDir query params. The repository
// whitelists sortBy against its known columns, so an unrecognized value simply
// falls back to the query's default ordering rather than erroring.
func ParseSort(r *http.Request) commandlog.SortSpec {
	q := r.URL.Query()
	return commandlog.SortSpec{
		By:  q.Get("sortBy"),
		Dir: q.Get("sortDir"),
	}
}
