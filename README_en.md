# Maintenance Dashboard

Web dashboard to visualise and analyse SQL Server maintenance metrics recorded by [Ola Hallengren's Maintenance Solution](https://ola.hallengren.com/) in the `MaintenanceDB.dbo.CommandLog` table.

## Table of contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick start with pre-built binary](#quick-start-with-pre-built-binary)
- [Build from source](#build-from-source)
- [Configuration](#configuration)
- [Encrypting credentials](#encrypting-credentials)
- [Running](#running)
- [Available pages](#available-pages)
- [REST API](#rest-api)

---

## Features

- **Servers** — reachability status of all configured SQL Server connections
- **Dashboard** — global summary: total commands, errors, average duration, counts by type
- **Top Fragmented Indexes** — indexes ordered by number of maintenance operations, with fragmentation percentage extracted from the `ExtendedInfo` XML column
- **Most Modified Statistics** — objects with the most statistics updates, including page count and modification counter from the `ExtendedInfo` XML column
- **Operations Per Batch** — operation count per period and command type (chart + table)
- **Maintenance Summary** — total duration per command type over time (line chart + table)
- **Backup Overview** — backup duration and error count per database (bar charts)
- **Maintenance Errors** — failed operations: timeline chart, summary by type, detail with SQL Server error number and message
- **Long Running Operations** — operations exceeding a configurable duration threshold, with a chart of maximum duration over time and a table of the individual slowest operations

All pages (except Servers) support filtering by database, date range, and granularity (daily / weekly / monthly). When multiple servers are configured, a server selector is available on every page.

---

## Prerequisites

### Database

- **SQL Server** 2016 or later (any edition)
- [Ola Hallengren's Maintenance Solution](https://ola.hallengren.com/) must be installed and must have created the `dbo.CommandLog` table in the `MaintenanceDB` database (or whichever database is set in the `database.name` field)
- The configured SQL login must have at least `SELECT` permission on the `CommandLog` table

### Pre-built binary

No additional dependencies. The binaries include the frontend already embedded.

### Build from source

- **Go** 1.26 or later — [golang.org/dl](https://golang.org/dl/)
- **Node.js** 18 or later with **npm** — [nodejs.org](https://nodejs.org/)
- **make** — available by default on macOS/Linux; on Windows install via [Chocolatey](https://chocolatey.org/) (`choco install make`) or use Git Bash

---

## Quick start with pre-built binary

Pre-built binaries are located in the `_releases/` folder.

| Platform | File |
|---|---|
| Windows (64-bit) | `_releases/maintenance-dashboard-windows-amd64.exe` |
| macOS Intel | `_releases/maintenance-dashboard-darwin-amd64` |
| macOS Apple Silicon | `_releases/maintenance-dashboard-darwin-arm64` |
| Linux (64-bit) | `_releases/maintenance-dashboard-linux-amd64` |

### Windows

1. Copy `maintenance-dashboard-windows-amd64.exe` to a folder of your choice
2. In that same folder, create a `configs/` sub-folder and copy `configs/config.example.yaml` into it, renaming it `config.yaml`
3. Edit `config.yaml` with your SQL Server parameters (see [Configuration](#configuration))
4. Run the program:

```powershell
.\maintenance-dashboard-windows-amd64.exe
```

### macOS

1. Copy the appropriate binary (`darwin-amd64` for Intel, `darwin-arm64` for Apple Silicon) to a folder of your choice
2. In that same folder, create a `configs/` sub-folder and copy `configs/config.example.yaml` into it, renaming it `config.yaml`
3. Edit `config.yaml` with your SQL Server parameters
4. Make the file executable and run it:

```bash
chmod +x ./maintenance-dashboard-darwin-arm64
./maintenance-dashboard-darwin-arm64
```

### Linux

1. Copy `maintenance-dashboard-linux-amd64` to a folder of your choice
2. In that same folder, create a `configs/` sub-folder and copy `configs/config.example.yaml` into it, renaming it `config.yaml`
3. Edit `config.yaml` with your SQL Server parameters
4. Make the file executable and run it:

```bash
chmod +x ./maintenance-dashboard-linux-amd64
./maintenance-dashboard-linux-amd64
```

Once started, open your browser at `http://localhost:8080` (or the configured port).

---

## Build from source

### 1. Clone the repository

```bash
git clone <repository-url>
cd maintenance-dashboard
```

### 2. Configure

```bash
cp configs/config.example.yaml configs/config.yaml
# edit configs/config.yaml with your SQL Server parameters
```

### 3. Build (with Makefile)

Build the frontend and the Go binary for the current platform in one step:

```bash
make
```

To produce release binaries for all platforms at once:

```bash
make release
# outputs to _releases/
```

Available targets:

| Target | Output |
|---|---|
| `make` | Frontend + binary for current platform |
| `make build-windows` | `_releases/maintenance-dashboard-windows-amd64.exe` |
| `make build-mac` | `_releases/maintenance-dashboard-darwin-{amd64,arm64}` |
| `make build-linux` | `_releases/maintenance-dashboard-linux-amd64` |
| `make release` | Frontend + all platforms |

Run `make help` for the full list of targets.

### 3b. Build manually (without make)

<details>
<summary>Manual build steps</summary>

**Frontend:**

```bash
cd web/frontend
npm install
npm run build
cd ../..
```

`npm run build` outputs static files to `internal/infra/assets/dist/`, which are embedded into the Go binary via `//go:embed`.

**Backend (current platform):**

```bash
go build -o maintenance-dashboard ./cmd/server
```

On Windows:

```powershell
go build -o maintenance-dashboard.exe ./cmd/server
```

**Cross-compilation:**

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o _releases/maintenance-dashboard-windows-amd64.exe ./cmd/server

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o _releases/maintenance-dashboard-darwin-arm64 ./cmd/server

# Linux
GOOS=linux GOARCH=amd64 go build -o _releases/maintenance-dashboard-linux-amd64 ./cmd/server
```

</details>

### 4. Run

```bash
./maintenance-dashboard
```

### Development mode (with hot reload)

To work on the frontend without recompiling Go on every change, start the two processes separately:

```bash
# Terminal 1 — Go backend
make run          # or: go run ./cmd/server

# Terminal 2 — Vite frontend with proxy to the backend
make run-frontend # or: cd web/frontend && npm run dev
```

The Vite server runs on `http://localhost:5173` and automatically proxies all `/api/*` calls to the Go backend on `http://localhost:8080`.

---

## Configuration

The program loads configuration in the following order of precedence (lowest to highest):

1. Internal default values
2. `configs/config.yaml` file
3. Environment variables (or a `.env` file in the same folder as the binary)

Environment variables always override values in the YAML file. The `.env` file is only loaded for variables not already present in the operating system environment.

### Full configuration file

```yaml
app:
  name: maintenance-dashboard   # process name (used in logs)
  env: production               # development | production
  host: 0.0.0.0                 # listen address (0.0.0.0 = all interfaces)
  port: 8080                    # HTTP port
  read_timeout_seconds: 15
  write_timeout_seconds: 15
  idle_timeout_seconds: 60

auth:
  enabled: true                 # false = no authentication
  username: admin
  password: change-me           # plain-text or encrypted (enc:...) — change before going to production

# Multiple servers — each entry is selectable from the UI dropdown.
# mode: sql = SQL Server authentication; integrated = Windows Authentication (Windows hosts only)
servers:
  - name: production            # display name shown in the UI
    database:
      mode: sql
      host: prod-server         # hostname or IP address of the SQL Server
      port: 1433                # port (SQL Server default: 1433)
      instance: ""              # named instance (e.g. "SQLEXPRESS"); leave empty for default
      name: MaintenanceDB       # database that contains CommandLog
      username: sa              # SQL login (ignored when mode: integrated) — plain-text or enc:...
      password: your-password   # SQL password (ignored when mode: integrated) — plain-text or enc:...
      encrypt: false            # true = encrypt the connection with TLS
      trust_server_certificate: true # true = accept self-signed certificates
      connection_timeout_seconds: 10
      max_open_conns: 20        # maximum open connections in the pool
      max_idle_conns: 10        # maximum idle connections in the pool
      conn_max_lifetime_minutes: 30 # maximum lifetime of a pooled connection
  - name: staging               # second server — add as many entries as needed
    database:
      mode: sql
      host: staging-server
      port: 1433
      instance: ""
      name: MaintenanceDB
      username: sa
      password: your-password
      encrypt: false
      trust_server_certificate: true
      connection_timeout_seconds: 10
      max_open_conns: 10
      max_idle_conns: 5
      conn_max_lifetime_minutes: 30

# Legacy single-server format (still supported — auto-converted to a "default" servers entry):
# database:
#   mode: sql
#   host: localhost
#   ...

cache:
  enabled: true                 # false = disable cache (every request hits the DB)
  dashboard_ttl_seconds: 30     # TTL for dashboard data
  detail_ttl_seconds: 60        # TTL for detail pages
  filters_ttl_seconds: 300      # TTL for filter options (databases, types, etc.)
  default_ttl_seconds: 60       # fallback TTL
  cleanup_interval_seconds: 60  # how often expired entries are purged

ui:
  default_language: en          # en | it
  supported_languages:
    - en
    - it
```

### Environment variables

Every `app`, `auth`, `cache`, and `ui` YAML field has a corresponding environment variable. Variables override the YAML value.

| Variable | YAML field | Example |
|---|---|---|
| `APP_PORT` | `app.port` | `8080` |
| `APP_HOST` | `app.host` | `0.0.0.0` |
| `APP_ENV` | `app.env` | `production` |
| `AUTH_ENABLED` | `auth.enabled` | `true` |
| `AUTH_USERNAME` | `auth.username` | `admin` |
| `AUTH_PASSWORD` | `auth.password` | `s3cr3t` |
| `ENCRYPTION_KEY` | — | 32-byte key as 64 hex chars or base64 (enables credential encryption) |
| `CACHE_ENABLED` | `cache.enabled` | `true` |
| `CACHE_DEFAULT_TTL_SECONDS` | `cache.default_ttl_seconds` | `60` |
| `CACHE_DASHBOARD_TTL_SECONDS` | `cache.dashboard_ttl_seconds` | `30` |
| `CACHE_DETAIL_TTL_SECONDS` | `cache.detail_ttl_seconds` | `60` |
| `CACHE_FILTERS_TTL_SECONDS` | `cache.filters_ttl_seconds` | `300` |
| `CACHE_CLEANUP_INTERVAL_SECONDS` | `cache.cleanup_interval_seconds` | `60` |
| `UI_DEFAULT_LANGUAGE` | `ui.default_language` | `it` |
| `UI_SUPPORTED_LANGUAGES` | `ui.supported_languages` | `en,it` |

Server connection parameters (host, credentials, etc.) must be set in `configs/config.yaml` under the `servers:` list; they do not map to environment variables.

### Example `.env` file

```dotenv
AUTH_PASSWORD=my-secure-password
ENCRYPTION_KEY=0102...1e1f   # 64 hex chars; keep this out of config.yaml
```

### Windows Authentication

To use integrated Windows Authentication (Windows hosts only), set `mode: integrated` in the server entry:

```yaml
servers:
  - name: production
    database:
      mode: integrated
      host: localhost
      name: MaintenanceDB
      # username and password are ignored
```

### Named instance

```yaml
servers:
  - name: production
    database:
      host: MYSERVER
      instance: SQLEXPRESS   # connects to MYSERVER\SQLEXPRESS
      port: 0                # port is ignored when an instance name is specified
```

### TLS connection

```yaml
servers:
  - name: production
    database:
      encrypt: true
      trust_server_certificate: false  # false = verify the certificate (recommended in production)
```

---

## Encrypting credentials

Sensitive fields in `config.yaml` — `auth.password` and `servers[].database.username` / `password` — can be stored encrypted instead of in plain text. Encrypted values use the format `enc:<base64>` and are decrypted at startup with AES-256-GCM.

Plain-text values continue to work as before; encryption is opt-in per field.

### Setup

**1. Generate a 32-byte key**

```bash
openssl rand -hex 32
# example output: 5f3b...e4a1  (64 hex characters)
```

Store the key in `.env` or your secret manager — **never** put it in `config.yaml`.

**.env**
```dotenv
ENCRYPTION_KEY=5f3b...e4a1
```

**2. Encrypt a value**

```bash
ENCRYPTION_KEY=5f3b...e4a1 ./maintenance-dashboard encrypt "my-sql-password"
# → enc:6T9abc...==
```

On Windows (PowerShell):

```powershell
$env:ENCRYPTION_KEY = "5f3b...e4a1"
.\maintenance-dashboard-windows-amd64.exe encrypt "my-sql-password"
```

**3. Paste the result into `config.yaml`**

```yaml
auth:
  password: enc:6T9abc...==

servers:
  - name: production
    database:
      mode: sql
      username: enc:Xp1...==
      password: enc:6T9abc...==
```

### Notes

- If `ENCRYPTION_KEY` is not set, all values are read as plain text — nothing changes
- If a field contains `enc:...` but `ENCRYPTION_KEY` is absent, the server fails to start with a clear error
- Each call to `encrypt` produces a different ciphertext (random nonce) for the same plaintext — this is expected and secure
- The same key must be used for both encrypting and running the server

---

## Running

```bash
./maintenance-dashboard
```

The program prints the listen address on startup:

```
2024/01/15 10:00:00 starting maintenance-dashboard on 0.0.0.0:8080
```

Open your browser at `http://<host>:<port>`. If authentication is enabled the browser will show an HTTP Basic login prompt.

### Health check

The `/api/meta/health` endpoint requires no authentication and returns `200 OK` when the server is running:

```bash
curl http://localhost:8080/api/meta/health
```

---

## Available pages

| Path | Page | Main filters |
|---|---|---|
| `/servers` | Servers | — |
| `/` | Dashboard | Server, database, dates, command type, errors only |
| `/fragmented-indexes` | Top Fragmented Indexes | Server, database, dates, top N |
| `/modified-statistics` | Most Modified Statistics | Server, database, dates, top N |
| `/operations-per-batch` | Operations Per Batch | Server, database, dates, granularity |
| `/maintenance-summary` | Maintenance Summary | Server, database, dates, granularity |
| `/backup-overview` | Backup Overview | Server, database, dates, granularity |
| `/maintenance-errors` | Maintenance Errors | Server, database, command type, dates, granularity |
| `/long-running-operations` | Long Running Operations | Server, database, command type, dates, granularity, min duration |

---

## REST API

All endpoints (except `/api/meta/health`) require HTTP Basic Auth when `auth.enabled: true`.

| Method | Endpoint | Query parameters |
|---|---|---|
| `GET` | `/api/meta/health` | — |
| `GET` | `/api/meta/servers` | — |
| `GET` | `/api/meta/server-status` | — |
| `GET` | `/api/meta/filters` | `server` |
| `GET` | `/api/dashboard/summary` | standard filters |
| `GET` | `/api/statistics/most-modified` | standard filters + `limit` (1–500, default 50) |
| `GET` | `/api/indexes/top-fragmented` | standard filters + `limit` (1–500, default 50) |
| `GET` | `/api/maintenance/summary` | standard filters + `granularity` (day/week/month) |
| `GET` | `/api/operations/per-batch` | standard filters + `granularity` |
| `GET` | `/api/backup/report` | standard filters + `granularity` |
| `GET` | `/api/errors/report` | standard filters + `granularity` + `limit` (1–500, default 100) |
| `GET` | `/api/longrunning/report` | standard filters + `granularity` + `minDuration` (seconds, default 300) + `limit` |

**Standard filters** — query parameters shared by all data endpoints:

| Parameter | Type | Description |
|---|---|---|
| `server` | string | Target server name (as defined in `servers[].name`); defaults to first server |
| `database` | string | Filter by database name |
| `dateFrom` | `YYYY-MM-DD` | Start date (inclusive) |
| `dateTo` | `YYYY-MM-DD` | End date (inclusive) |
| `commandType` | string | Filter by command type (e.g. `ALTER_INDEX`, `UPDATE_STATISTICS`) |
| `schema` | string | Filter by schema name |
| `object` | string | Filter by object name |
| `onlyErrors` | `true`/`false` | Show only failed operations |

All responses use the format:

```json
{ "data": { ... } }
```

Errors return:

```json
{ "error": { "code": "...", "message": "..." } }
```
