package httpx

import (
	"net/http"
	"strconv"

	applongrunning "maintenance-dashboard/internal/app/longrunning"
)

type LongRunningHandler struct{ Services map[string]*applongrunning.Service }

func (h LongRunningHandler) Report(w http.ResponseWriter, r *http.Request) {
	svc, err := resolveServer(h.Services, r)
	if err != nil {
		Error(w, http.StatusBadRequest, "unknown_server", err.Error())
		return
	}
	filters, err := ParseFilters(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid_filters", err.Error())
		return
	}

	granularity := r.URL.Query().Get("granularity")
	switch granularity {
	case "day", "week", "month":
		// valid
	case "":
		granularity = "day"
	default:
		Error(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	minDuration := 300
	if v := r.URL.Query().Get("minDuration"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			Error(w, http.StatusBadRequest, "invalid_min_duration", "minDuration must be a non-negative integer (seconds)")
			return
		}
		minDuration = n
	}

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 500 {
			Error(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 500")
			return
		}
		limit = n
	}

	data, err := svc.GetLongRunningReport(r.Context(), filters, granularity, minDuration, limit)
	if err != nil {
		Error(w, http.StatusInternalServerError, "long_running_report_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
