package httpx

import (
	"net/http"

	appbackup "maintenance-dashboard/internal/app/backup"
)

type BackupHandler struct{ Service *appbackup.Service }

func (h BackupHandler) Report(w http.ResponseWriter, r *http.Request) {
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
	data, err := h.Service.GetBackupReport(r.Context(), filters, granularity)
	if err != nil {
		Error(w, http.StatusInternalServerError, "backup_report_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
