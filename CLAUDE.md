# CLAUDE.md — Maintenance Dashboard

## Project Overview

Full-stack dashboard for SQL Server maintenance analytics. Reads data from `MaintenanceDB.dbo.CommandLog` and visualizes index maintenance operations, statistics updates, query performance, and error tracking.

**Stack:**
- Backend: Go 1.26, standard library HTTP server
- Frontend: React 18 + TypeScript + Vite
- Database: Microsoft SQL Server (`go-mssqldb`)
- Styling: Tailwind CSS 3.4
- State/Fetching: TanStack React Query v5
- Routing: React Router v6
- Charts: Recharts 3.8
- i18n: i18next (English + Italian)

---

## Directory Structure

```
cmd/server/           # Entry point (main.go)
internal/
  app/                # Service layer (business logic + caching)
    dashboard/
    indexes/
    maintenance/
    meta/
    statistics/
    shared/
  domain/commandlog/  # Repository interfaces and domain models
  bootstrap/app.go    # DI container — wires all dependencies
  infra/
    assets/           # Embedded React build (go:embed)
    auth/             # HTTP Basic Auth middleware
    cache/            # In-memory TTL cache
    config/           # YAML + env config loading and validation
    db/               # SQL Server connection setup
    http/             # Handlers and router
    logging/
    repository/       # Concrete SQL implementations
configs/
  config.example.yaml
  config.yaml         # Runtime config (not committed)
.env                  # Environment variable overrides
web/frontend/
  src/
    api/              # HTTP client wrappers
    app/              # App layout and providers
    components/       # Shared UI components
    features/         # Feature pages (dashboard, indexes, ...)
    hooks/            # Custom React hooks
    i18n/             # en.json / it.json translation files
    types/            # TypeScript interfaces
  package.json
  vite.config.ts
  tailwind.config.ts
_releases/            # Pre-built binaries (Mac / Windows)
```

---

## Commands

### Makefile (preferred)
```bash
make                   # Build frontend + Go binary for current platform
make build             # Go binary only (current platform)
make build-frontend    # React build → internal/infra/assets/dist
make build-windows     # Cross-compile → _releases/maintenance-dashboard-windows-amd64.exe
make build-mac         # Cross-compile → _releases/maintenance-dashboard-darwin-{amd64,arm64}
make build-linux       # Cross-compile → _releases/maintenance-dashboard-linux-amd64
make release           # Frontend + all platform binaries
make run               # Run Go dev server
make run-frontend      # Run Vite dev server
make clean             # Remove build artifacts
make help              # Show all targets
```

### Backend (manual)
```bash
go run ./cmd/server          # Run development server
go build ./cmd/server        # Build binary
```

### Frontend (from web/frontend/)
```bash
npm run dev                  # Vite dev server (proxies to Go backend)
npm run build                # Production build → internal/infra/assets/dist
npm run preview              # Preview production build
```

### Full build flow
1. `cd web/frontend && npm run build` — outputs to `internal/infra/assets/dist`
2. `go build ./cmd/server` — Go embeds the dist folder via `//go:embed all:dist`

Or simply: `make` (runs both steps in order)

---

## Configuration

Two-tier config: YAML file first, environment variables override.

**Config file:** `configs/config.yaml` (copy from `configs/config.example.yaml`)

**Key sections:**

```yaml
app:
  host, port, env
  read/write/idle timeout seconds

auth:
  enabled, username, password

database:
  mode: sql | integrated
  host, port, instance, name
  username, password
  encrypt, trust_server_certificate

cache:
  enabled
  default_ttl_seconds, dashboard_ttl_seconds, detail_ttl_seconds, filters_ttl_seconds

ui:
  default_language: en | it
  supported_languages
```

**Environment variable examples:**
```
APP_PORT, APP_HOST
AUTH_USERNAME, AUTH_PASSWORD
DATABASE_MODE, DATABASE_HOST, DATABASE_PORT, DATABASE_NAME, DATABASE_USERNAME, DATABASE_PASSWORD
CACHE_ENABLED, CACHE_DEFAULT_TTL_SECONDS
UI_DEFAULT_LANGUAGE
```

The `.env` file is loaded at startup and only sets variables not already present in the OS environment.

---

## Architecture

### Backend — Clean Architecture (layered)

```
HTTP Handler  →  Service (app/)  →  Repository Interface (domain/)
                     ↓                         ↓
               Cache (infra/cache)   SQL Impl (infra/repository)
```

- **Domain** defines repository interfaces (`MetaRepository`, `DashboardRepository`, etc.)
- **Infrastructure** implements them with raw SQL queries
- **Services** depend on interfaces only, never concrete types
- **bootstrap/app.go** constructs the full dependency graph
- All services accept a common `QueryFilters` struct; URL params → domain filters → SQL WHERE clauses

### Caching
- All service methods cache results keyed by SHA1 hash of filters
- TTLs configurable per data type; background cleanup goroutine
- Can be disabled entirely via config

### Middleware
- Recovery (panic handler)
- 30-second global request timeout
- Optional HTTP Basic Auth

### Error handling
- Wrapped errors: `fmt.Errorf("context: %w", err)`
- JSON error responses: `{"error": {"code": "...", "message": "..."}}`

### Frontend — Feature-based
- Each page isolated under `src/features/<name>/`
- Shared UI in `src/components/`
- React Query manages all server state
- `src/api/` wraps raw fetch calls

---

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/meta/health` | No | Health check |
| GET | `/api/meta/filters` | Yes | Available filter options |
| GET | `/api/dashboard/summary` | Yes | Dashboard metrics |
| GET | `/api/statistics/most-modified` | Yes | Most updated statistics |
| GET | `/api/indexes/top-fragmented` | Yes | Fragmented indexes |
| GET | `/api/maintenance/summary` | Yes | Maintenance report |

Common query parameters: `dateFrom`, `dateTo`, `database`, `schema`, `objectName`, `commandType`, `hasErrors`, `limit`.

---

## Implemented Pages

| Route | Page |
|-------|------|
| `/` | Dashboard — summary cards with totals and filter options |
| `/modified-statistics` | Most Modified Statistics — paginated, limit 10–500 |
| `/fragmented-indexes` | Top Fragmented Indexes |
| `/maintenance-summary` | Maintenance Summary — timeline and aggregates |

**Placeholder routes (not yet implemented):** Operations Per Batch, Maintenance Errors, Long Running Operations.

---

## i18n

- Supported: `en` (English), `it` (Italian)
- Translation files: `web/frontend/src/i18n/en.json`, `it.json`
- Default language configured in `ui.default_language`

---

## Testing

No tests exist yet. The architecture is test-ready: all services depend on interfaces, making unit testing straightforward with mock implementations. When adding tests:
- Go: `*_test.go` files alongside the packages; use interface mocks for repositories
- Frontend: add Vitest (compatible with Vite setup)

---

## Notes

- Code is in English; some comments and docs may be in Italian
- `.idea/` present — JetBrains IDE (GoLand / WebStorm)
- `_releases/` contains pre-built binaries, not part of the main build process
- A `Makefile` is provided at the root for all common build tasks (see [Commands](#commands))
