package httpx

import (
	"net/http"

	appmaintenance "maintenance-dashboard/internal/app/maintenance"
)

type MaintenanceHandler struct{ Service *appmaintenance.Service }

func (h MaintenanceHandler) Summary(w http.ResponseWriter, r *http.Request) {
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
	data, err := h.Service.GetMaintenanceReport(r.Context(), filters, granularity)
	if err != nil {
		Error(w, http.StatusInternalServerError, "maintenance_summary_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
