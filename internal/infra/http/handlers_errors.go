package httpx

import (
	"net/http"
	"strconv"

	apperrors "maintenance-dashboard/internal/app/errors"
)

type MaintenanceErrorsHandler struct{ Services map[string]*apperrors.Service }

func (h MaintenanceErrorsHandler) Report(w http.ResponseWriter, r *http.Request) {
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

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 500 {
			Error(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 500")
			return
		}
		limit = n
	}

	data, err := svc.GetMaintenanceErrorsReport(r.Context(), filters, granularity, limit)
	if err != nil {
		Error(w, http.StatusInternalServerError, "maintenance_errors_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
