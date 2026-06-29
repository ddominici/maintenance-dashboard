package httpx

import (
	"net/http"
	"strconv"

	appindexes "maintenance-dashboard/internal/app/indexes"
)

type IndexesHandler struct{ Services map[string]*appindexes.Service }

func (h IndexesHandler) TopFragmented(w http.ResponseWriter, r *http.Request) {
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
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 500 {
			Error(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 500")
			return
		}
		limit = n
	}
	data, err := svc.GetTopFragmentedIndexes(r.Context(), filters, limit, ParseSort(r))
	if err != nil {
		Error(w, http.StatusInternalServerError, "indexes_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
