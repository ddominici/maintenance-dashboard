package httpx

import (
	"net/http"

	appdashboard "maintenance-dashboard/internal/app/dashboard"
)

type DashboardHandler struct{ Services map[string]*appdashboard.Service }

func (h DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
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
	data, err := svc.GetSummary(r.Context(), filters)
	if err != nil {
		Error(w, http.StatusInternalServerError, "dashboard_summary_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
