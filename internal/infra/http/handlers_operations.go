package httpx

import (
	"net/http"

	appoperations "maintenance-dashboard/internal/app/operations"
)

type OperationsHandler struct{ Services map[string]*appoperations.Service }

func (h OperationsHandler) PerBatch(w http.ResponseWriter, r *http.Request) {
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
	data, err := svc.GetOperationsBatchReport(r.Context(), filters, granularity)
	if err != nil {
		Error(w, http.StatusInternalServerError, "operations_batch_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
