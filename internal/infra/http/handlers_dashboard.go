package httpx

import (
	"net/http"

	appdashboard "maintenance-dashboard/internal/app/dashboard"
)

type DashboardHandler struct{ Service *appdashboard.Service }

func (h DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	filters, err := ParseFilters(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid_filters", err.Error())
		return
	}
	data, err := h.Service.GetSummary(r.Context(), filters)
	if err != nil {
		Error(w, http.StatusInternalServerError, "dashboard_summary_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
