package httpx

import (
	"net/http"

	appmeta "maintenance-dashboard/internal/app/meta"
)

type MetaHandler struct{ Service *appmeta.Service }

func (h MetaHandler) Health(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h MetaHandler) Filters(w http.ResponseWriter, r *http.Request) {
	data, err := h.Service.GetFilterOptions(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "filters_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
