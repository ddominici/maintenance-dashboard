package httpx

import (
	"net/http"

	appmaintenance "maintenance-dashboard/internal/app/maintenance"
)

type MaintenanceHandler struct{ Services map[string]*appmaintenance.Service }

func (h MaintenanceHandler) Summary(w http.ResponseWriter, r *http.Request) {
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
	data, err := svc.GetMaintenanceReport(r.Context(), filters, granularity)
	if err != nil {
		Error(w, http.StatusInternalServerError, "maintenance_summary_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
