package httpx

import (
	"net/http"

	appoperations "maintenance-dashboard/internal/app/operations"
)

type OperationsHandler struct{ Service *appoperations.Service }

func (h OperationsHandler) PerBatch(w http.ResponseWriter, r *http.Request) {
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
	data, err := h.Service.GetOperationsBatchReport(r.Context(), filters, granularity)
	if err != nil {
		Error(w, http.StatusInternalServerError, "operations_batch_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
