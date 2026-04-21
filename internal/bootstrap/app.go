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
	dbs    map[string]*sql.DB
	server *http.Server
	logger *logging.Logger
}

func NewApp() (*App, error) {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		return nil, err
	}

	logger := logging.New()

	dbs := make(map[string]*sql.DB, len(cfg.Servers))
	for _, srv := range cfg.Servers {
		sqlDB, err := db.Connect(srv.Database)
		if err != nil {
			for _, d := range dbs {
				_ = d.Close()
			}
			return nil, fmt.Errorf("connect server %q: %w", srv.Name, err)
		}
		dbs[srv.Name] = sqlDB
	}

	var memCache cache.Cache
	if cfg.Cache.Enabled {
		memCache = cache.NewMemoryCache(time.Duration(cfg.Cache.CleanupIntervalSeconds) * time.Second)
	}

	metaServices := make(map[string]*appmeta.Service, len(cfg.Servers))
	dashServices := make(map[string]*appdashboard.Service, len(cfg.Servers))
	statsServices := make(map[string]*appstatistics.Service, len(cfg.Servers))
	indexesServices := make(map[string]*appindexes.Service, len(cfg.Servers))
	maintenanceServices := make(map[string]*appmaintenance.Service, len(cfg.Servers))
	backupServices := make(map[string]*appbackup.Service, len(cfg.Servers))
	operationsServices := make(map[string]*appoperations.Service, len(cfg.Servers))
	longRunningServices := make(map[string]*applongrunning.Service, len(cfg.Servers))
	errorsServices := make(map[string]*apperrors.Service, len(cfg.Servers))

	for _, srv := range cfg.Servers {
		repo := repository.NewCommandLogRepository(dbs[srv.Name])
		metaServices[srv.Name] = appmeta.NewService(repo, memCache, time.Duration(cfg.Cache.FiltersTTLSeconds)*time.Second)
		dashServices[srv.Name] = appdashboard.NewService(repo, memCache, time.Duration(cfg.Cache.DashboardTTLSeconds)*time.Second)
		statsServices[srv.Name] = appstatistics.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
		indexesServices[srv.Name] = appindexes.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
		maintenanceServices[srv.Name] = appmaintenance.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
		backupServices[srv.Name] = appbackup.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
		operationsServices[srv.Name] = appoperations.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
		longRunningServices[srv.Name] = applongrunning.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
		errorsServices[srv.Name] = apperrors.NewService(repo, memCache, time.Duration(cfg.Cache.DetailTTLSeconds)*time.Second)
	}

	serverNames := make([]string, len(cfg.Servers))
	serverInfos := make([]httpx.ServerInfo, len(cfg.Servers))
	for i, s := range cfg.Servers {
		serverNames[i] = s.Name
		serverInfos[i] = httpx.ServerInfo{Name: s.Name, Host: s.Database.Host, DB: dbs[s.Name]}
	}

	router, err := httpx.NewRouter(cfg, httpx.Handlers{
		Meta:              httpx.MetaHandler{Services: metaServices, ServerNames: serverNames, ServerInfos: serverInfos},
		Dashboard:         httpx.DashboardHandler{Services: dashServices},
		Statistics:        httpx.StatisticsHandler{Services: statsServices},
		Indexes:           httpx.IndexesHandler{Services: indexesServices},
		Maintenance:       httpx.MaintenanceHandler{Services: maintenanceServices},
		Backup:            httpx.BackupHandler{Services: backupServices},
		Operations:        httpx.OperationsHandler{Services: operationsServices},
		LongRunning:       httpx.LongRunningHandler{Services: longRunningServices},
		MaintenanceErrors: httpx.MaintenanceErrorsHandler{Services: errorsServices},
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

	return &App{cfg: cfg, dbs: dbs, server: srv, logger: logger}, nil
}

func (a *App) Run() error {
	a.logger.Infof("starting %s on %s", a.cfg.App.Name, a.server.Addr)
	return a.server.ListenAndServe()
}
