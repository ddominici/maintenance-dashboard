package httpx

import (
	"io/fs"
	"net/http"
	"time"

	"maintenance-dashboard/internal/infra/assets"
	"maintenance-dashboard/internal/infra/auth"
	"maintenance-dashboard/internal/infra/config"
)

type Handlers struct {
	Meta        MetaHandler
	Dashboard   DashboardHandler
	Statistics  StatisticsHandler
	Indexes     IndexesHandler
	Maintenance MaintenanceHandler
	Backup      BackupHandler
	Operations  OperationsHandler
	LongRunning       LongRunningHandler
	MaintenanceErrors MaintenanceErrorsHandler
}

func NewRouter(cfg config.Config, h Handlers) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/meta/health", h.Meta.Health)

	protected := http.NewServeMux()
	protected.HandleFunc("/api/meta/filters", h.Meta.Filters)
	protected.HandleFunc("/api/dashboard/summary", h.Dashboard.Summary)
	protected.HandleFunc("/api/statistics/most-modified", h.Statistics.MostModified)
	protected.HandleFunc("/api/indexes/top-fragmented", h.Indexes.TopFragmented)
	protected.HandleFunc("/api/maintenance/summary", h.Maintenance.Summary)
	protected.HandleFunc("/api/backup/report", h.Backup.Report)
	protected.HandleFunc("/api/operations/per-batch", h.Operations.PerBatch)
	protected.HandleFunc("/api/longrunning/report", h.LongRunning.Report)
	protected.HandleFunc("/api/errors/report", h.MaintenanceErrors.Report)

	staticFS, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		return nil, err
	}
	protected.Handle("/", spaHandler(staticFS))

	var root http.Handler = protected
	if cfg.Auth.Enabled {
		root = auth.Basic(cfg.Auth.Username, cfg.Auth.Password, protected)
	}
	mux.Handle("/", root)

	return withMiddleware(mux), nil
}

func withMiddleware(next http.Handler) http.Handler {
	return recoverer(timeoutMiddleware(30 * time.Second)(next))
}

func timeoutMiddleware(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return http.TimeoutHandler(next, d, "request timeout") }
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				Error(w, http.StatusInternalServerError, "internal_error", "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For paths other than root, check if the file actually exists.
		// If not, fall back to "/" so FileServer serves index.html via its
		// built-in directory-index logic — without triggering the redirect
		// loop that happens when we set the path to "/index.html" directly
		// (FileServer redirects "/index.html" → "/" to avoid duplicate content).
		if r.URL.Path != "/" {
			if _, err := fs.Stat(fsys, r.URL.Path[1:]); err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
