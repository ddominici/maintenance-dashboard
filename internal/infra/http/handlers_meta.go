package httpx

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	appmeta "maintenance-dashboard/internal/app/meta"
)

type ServerInfo struct {
	Name string
	Host string
	DB   *sql.DB
}

type MetaHandler struct {
	Services    map[string]*appmeta.Service
	ServerNames []string
	ServerInfos []ServerInfo
}

func (h MetaHandler) Health(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h MetaHandler) Servers(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]any{"data": h.ServerNames})
}

func (h MetaHandler) ServerStatus(w http.ResponseWriter, r *http.Request) {
	type item struct {
		Name      string `json:"name"`
		Host      string `json:"host"`
		Reachable bool   `json:"reachable"`
	}

	result := make([]item, len(h.ServerInfos))
	var wg sync.WaitGroup
	for i, info := range h.ServerInfos {
		wg.Add(1)
		go func(idx int, si ServerInfo) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			result[idx] = item{
				Name:      si.Name,
				Host:      si.Host,
				Reachable: si.DB.PingContext(ctx) == nil,
			}
		}(i, info)
	}
	wg.Wait()
	JSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h MetaHandler) Filters(w http.ResponseWriter, r *http.Request) {
	svc, err := resolveServer(h.Services, r)
	if err != nil {
		Error(w, http.StatusBadRequest, "unknown_server", err.Error())
		return
	}
	data, err := svc.GetFilterOptions(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "filters_failed", err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"data": data})
}
