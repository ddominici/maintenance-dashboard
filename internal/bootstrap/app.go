package bootstrap

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	appbackup "maintenance-dashboard/internal/app/backup"
	appdashboard "maintenance-dashboard/internal/app/dashboard"
	apperrors "maintenance-dashboard/internal/app/errors"
	appindexes "maintenance-dashboard/internal/app/indexes"
	applongrunning "maintenance-dashboard/internal/app/longrunning"
	appoperations "maintenance-dashboard/internal/app/operations"
	appmaintenance "maintenance-dashboard/internal/app/maintenance"
	appmeta "maintenance-dashboard/internal/app/meta"
	appstatistics "maintenance-dashboard/internal/app/statistics"
	"maintenance-dashboard/internal/infra/cache"
	"maintenance-dashboard/internal/infra/config"
	"maintenance-dashboard/internal/infra/db"
	httpx "maintenance-dashboard/internal/infra/http"
	"maintenance-dashboard/internal/infra/logging"
	"maintenance-dashboard/internal/infra/repository"
)

type App struct {
	cfg    config.Config
	db     *sql.DB
	server *http.Server
	logger *logging.Logger
}

func NewApp() (*App, error) {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		return nil, err
	}

	logger := logging.New()
	sqlDB, err := db.Connect(cfg.Database)
	if err != nil {
		return nil, err
	}

	var memCache cache.Cache
	if cfg.Cache.Enabled {
		memCache = cache.NewMemoryCache(time.Duration(cfg.Cache.CleanupIntervalSeconds) * time.Second)
	}
	repo := repository.NewCommandLogRepository(sqlDB)

	metaSvc := appmeta.NewService(repo, memCache, time.Duration(cfg.Cache.FiltersTTLSeconds)*time.Second)
	dashSvc := appdashboard.NewService(repo, memCache, time.Duration(cfg.Cache.DashboardTTLSeconds)*time.Second)
	statsSvc := appstatistics.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
	indexesSvc := appindexes.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
	maintenanceSvc := appmaintenance.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
	backupSvc := appbackup.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
	operationsSvc := appoperations.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
	longRunningSvc := applongrunning.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
	errorsSvc := apperrors.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)

	router, err := httpx.NewRouter(cfg, httpx.Handlers{
		Meta:        httpx.MetaHandler{Service: metaSvc},
		Dashboard:   httpx.DashboardHandler{Service: dashSvc},
		Statistics:  httpx.StatisticsHandler{Service: statsSvc},
		Indexes:     httpx.IndexesHandler{Service: indexesSvc},
		Maintenance: httpx.MaintenanceHandler{Service: maintenanceSvc},
		Backup:      httpx.BackupHandler{Service: backupSvc},
		Operations:  httpx.OperationsHandler{Service: operationsSvc},
		LongRunning:       httpx.LongRunningHandler{Service: longRunningSvc},
		MaintenanceErrors: httpx.MaintenanceErrorsHandler{Service: errorsSvc},
	})
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.App.IdleTimeoutSeconds) * time.Second,
	}

	return &App{cfg: cfg, db: sqlDB, server: srv, logger: logger}, nil
}

func (a *App) Run() error {
	a.logger.Infof("starting %s on %s", a.cfg.App.Name, a.server.Addr)
	return a.server.ListenAndServe()
}
