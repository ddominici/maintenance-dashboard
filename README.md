# Maintenance Dashboard

Dashboard web per visualizzare e analizzare le metriche di manutenzione SQL Server registrate da [Ola Hallengren's Maintenance Solution](https://ola.hallengren.com/) nella tabella `MaintenanceDB.dbo.CommandLog`.

## Indice

- [Funzionalità](#funzionalità)
- [Prerequisiti](#prerequisiti)
- [Avvio rapido con binario precompilato](#avvio-rapido-con-binario-precompilato)
- [Build dal sorgente](#build-dal-sorgente)
- [Configurazione](#configurazione)
- [Avvio](#avvio)
- [Pagine disponibili](#pagine-disponibili)
- [API REST](#api-rest)

---

## Funzionalità

- **Servers** — stato di raggiungibilità di tutte le connessioni SQL Server configurate
- **Dashboard** — riepilogo globale: totale comandi, errori, durata media, conteggi per tipo
- **Top Fragmented Indexes** — indici ordinati per numero di operazioni di manutenzione, con percentuale di frammentazione estratta dall'XML `ExtendedInfo`
- **Most Modified Statistics** — oggetti con più aggiornamenti statistiche, con page count e modification counter dall'XML `ExtendedInfo`
- **Operations Per Batch** — conteggio operazioni per periodo e tipo di comando (grafico + tabella)
- **Maintenance Summary** — durata totale per tipo di comando nel tempo (grafico lineare + tabella)
- **Backup Overview** — durata e numero di errori dei backup per database (grafici a barre)
- **Maintenance Errors** — operazioni in errore: grafico temporale, riepilogo per tipo, dettaglio con numero e messaggio di errore SQL Server
- **Long Running Operations** — operazioni che superano una soglia di durata configurabile, con grafico della durata massima nel tempo e tabella delle singole operazioni più lente

Tutte le pagine (eccetto Servers) supportano filtri per database, intervallo di date e granularità (giornaliera / settimanale / mensile). Se sono configurati più server, un selettore server è disponibile su ogni pagina.

---

## Prerequisiti

### Database

- **SQL Server** 2016 o successivo (qualsiasi edizione)
- La [Maintenance Solution di Ola Hallengren](https://ola.hallengren.com/) deve essere installata e deve aver creato la tabella `dbo.CommandLog` nel database `MaintenanceDB` (o in qualsiasi database configurato nel campo `database.name`)
- L'utente SQL configurato deve avere almeno il permesso `SELECT` sulla tabella `CommandLog`

### Per il binario precompilato

Nessuna dipendenza aggiuntiva. I binari includono il frontend già integrato.

### Per build dal sorgente

- **Go** 1.26 o successivo — [golang.org/dl](https://golang.org/dl/)
- **Node.js** 18 o successivo con **npm** — [nodejs.org](https://nodejs.org/)
- **make** — disponibile di default su macOS/Linux; su Windows installare tramite [Chocolatey](https://chocolatey.org/) (`choco install make`) o usare Git Bash

---

## Avvio rapido con binario precompilato

I binari si trovano nella cartella `_releases/`.

| Piattaforma | File |
|---|---|
| Windows (64-bit) | `_releases/maintenance-dashboard-windows-amd64.exe` |
| macOS Intel | `_releases/maintenance-dashboard-darwin-amd64` |
| macOS Apple Silicon | `_releases/maintenance-dashboard-darwin-arm64` |
| Linux (64-bit) | `_releases/maintenance-dashboard-linux-amd64` |

### Windows

1. Copia `maintenance-dashboard-windows-amd64.exe` in una cartella di tua scelta
2. Nella stessa cartella, crea la sottocartella `configs/` e copiaci `configs/config.example.yaml` rinominandolo `config.yaml`
3. Modifica `config.yaml` con i parametri del tuo SQL Server (vedi [Configurazione](#configurazione))
4. Avvia il programma:

```powershell
.\maintenance-dashboard-windows-amd64.exe
```

### macOS

1. Copia il binario appropriato (`darwin-amd64` per Intel, `darwin-arm64` per Apple Silicon) in una cartella di tua scelta
2. Nella stessa cartella, crea la sottocartella `configs/` e copiaci `configs/config.example.yaml` rinominandolo `config.yaml`
3. Modifica `config.yaml` con i parametri del tuo SQL Server
4. Rendi il file eseguibile e avvialo:

```bash
chmod +x ./maintenance-dashboard-darwin-arm64
./maintenance-dashboard-darwin-arm64
```

### Linux

1. Copia `maintenance-dashboard-linux-amd64` in una cartella di tua scelta
2. Nella stessa cartella, crea la sottocartella `configs/` e copiaci `configs/config.example.yaml` rinominandolo `config.yaml`
3. Modifica `config.yaml` con i parametri del tuo SQL Server
4. Rendi il file eseguibile e avvialo:

```bash
chmod +x ./maintenance-dashboard-linux-amd64
./maintenance-dashboard-linux-amd64
```

Una volta avviato, apri il browser su `http://localhost:8080` (o la porta configurata).

---

## Build dal sorgente

### 1. Clona il repository

```bash
git clone <repository-url>
cd maintenance-dashboard
```

### 2. Configura

```bash
cp configs/config.example.yaml configs/config.yaml
# modifica configs/config.yaml con i parametri del tuo SQL Server
```

### 3. Build (con Makefile)

Frontend + binario Go per la piattaforma corrente in un solo comando:

```bash
make
```

Per generare i binari di release per tutte le piattaforme:

```bash
make release
# output in _releases/
```

Target disponibili:

| Target | Output |
|---|---|
| `make` | Frontend + binario per la piattaforma corrente |
| `make build-windows` | `_releases/maintenance-dashboard-windows-amd64.exe` |
| `make build-mac` | `_releases/maintenance-dashboard-darwin-{amd64,arm64}` |
| `make build-linux` | `_releases/maintenance-dashboard-linux-amd64` |
| `make release` | Frontend + tutte le piattaforme |

Esegui `make help` per la lista completa dei target.

### 3b. Build manuale (senza make)

<details>
<summary>Passi manuali</summary>

**Frontend:**

```bash
cd web/frontend
npm install
npm run build
cd ../..
```

Il comando `npm run build` genera i file statici in `internal/infra/assets/dist/`, che vengono poi incorporati nel binario Go tramite `//go:embed`.

**Backend (piattaforma corrente):**

```bash
go build -o maintenance-dashboard ./cmd/server
```

Su Windows:

```powershell
go build -o maintenance-dashboard.exe ./cmd/server
```

**Cross-compilazione:**

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o _releases/maintenance-dashboard-windows-amd64.exe ./cmd/server

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o _releases/maintenance-dashboard-darwin-arm64 ./cmd/server

# Linux
GOOS=linux GOARCH=amd64 go build -o _releases/maintenance-dashboard-linux-amd64 ./cmd/server
```

</details>

### 4. Avvio

```bash
./maintenance-dashboard
```

### Sviluppo (con hot reload)

Per lavorare sul frontend senza ricompilare Go a ogni modifica, avvia i due processi separatamente:

```bash
# Terminale 1 — backend Go
make run          # oppure: go run ./cmd/server

# Terminale 2 — frontend Vite con proxy verso il backend
make run-frontend # oppure: cd web/frontend && npm run dev
```

Il server Vite gira su `http://localhost:5173` e fa proxy automatico delle chiamate `/api/*` verso il backend Go su `http://localhost:8080`.

---

## Configurazione

Il programma carica la configurazione in questo ordine di precedenza (dal più basso al più alto):

1. Valori di default interni
2. File `configs/config.yaml`
3. Variabili d'ambiente (o file `.env` nella stessa cartella del binario)

Le variabili d'ambiente sovrascrivono sempre i valori nel file YAML. Il file `.env` viene caricato solo se la variabile non è già presente nell'ambiente del sistema operativo.

### File di configurazione completo

```yaml
app:
  name: maintenance-dashboard   # nome del processo (log)
  env: production               # development | production
  host: 0.0.0.0                 # indirizzo di ascolto (0.0.0.0 = tutte le interfacce)
  port: 8080                    # porta HTTP
  read_timeout_seconds: 15
  write_timeout_seconds: 15
  idle_timeout_seconds: 60

auth:
  enabled: true                 # false = nessuna autenticazione
  username: admin
  password: change-me           # cambiare prima della messa in produzione

# Più server — ogni voce è selezionabile dall'interfaccia utente.
# mode: sql = autenticazione SQL Server; integrated = Windows Authentication (solo Windows)
servers:
  - name: production            # nome visualizzato nell'interfaccia
    database:
      mode: sql
      host: localhost           # hostname o indirizzo IP del server SQL
      port: 1433                # porta (default SQL Server: 1433)
      instance: ""              # nome istanza named (es. "SQLEXPRESS"); lasciare vuoto per default
      name: MaintenanceDB       # database che contiene CommandLog
      username: sa              # utente SQL (ignorato se mode: integrated)
      password: your-password   # password SQL (ignorato se mode: integrated)
      encrypt: false            # true = cifra la connessione TLS
      trust_server_certificate: true # true = accetta certificati autofirmati
      connection_timeout_seconds: 10
      max_open_conns: 20        # connessioni aperte massime nel pool
      max_idle_conns: 10        # connessioni idle massime nel pool
      conn_max_lifetime_minutes: 30 # durata massima di una connessione nel pool

# Formato legacy a server singolo (ancora supportato — convertito automaticamente in servers[0] "default"):
# database:
#   mode: sql
#   host: localhost
#   ...

cache:
  enabled: true                 # false = disabilita la cache (ogni richiesta interroga il DB)
  dashboard_ttl_seconds: 30     # TTL per i dati della dashboard
  detail_ttl_seconds: 60        # TTL per le pagine di dettaglio
  filters_ttl_seconds: 300      # TTL per le opzioni di filtro (database, tipi, ecc.)
  default_ttl_seconds: 60       # TTL di fallback
  cleanup_interval_seconds: 60  # frequenza di pulizia delle voci scadute

ui:
  default_language: en          # en | it
  supported_languages:
    - en
    - it
```

### Variabili d'ambiente

I campi `app`, `auth`, `cache` e `ui` del file YAML corrispondono a variabili d'ambiente. Le variabili sovrascrivono il valore YAML.

| Variabile | Corrispondenza YAML | Esempio |
|---|---|---|
| `APP_PORT` | `app.port` | `8080` |
| `APP_HOST` | `app.host` | `0.0.0.0` |
| `APP_ENV` | `app.env` | `production` |
| `AUTH_ENABLED` | `auth.enabled` | `true` |
| `AUTH_USERNAME` | `auth.username` | `admin` |
| `AUTH_PASSWORD` | `auth.password` | `s3cr3t` |
| `CACHE_ENABLED` | `cache.enabled` | `true` |
| `CACHE_DEFAULT_TTL_SECONDS` | `cache.default_ttl_seconds` | `60` |
| `CACHE_DASHBOARD_TTL_SECONDS` | `cache.dashboard_ttl_seconds` | `30` |
| `CACHE_DETAIL_TTL_SECONDS` | `cache.detail_ttl_seconds` | `60` |
| `CACHE_FILTERS_TTL_SECONDS` | `cache.filters_ttl_seconds` | `300` |
| `CACHE_CLEANUP_INTERVAL_SECONDS` | `cache.cleanup_interval_seconds` | `60` |
| `UI_DEFAULT_LANGUAGE` | `ui.default_language` | `it` |
| `UI_SUPPORTED_LANGUAGES` | `ui.supported_languages` | `en,it` |

I parametri di connessione al server (host, credenziali, ecc.) devono essere impostati in `configs/config.yaml` nella lista `servers:`; non sono mappati a variabili d'ambiente.

### Esempio file `.env`

```dotenv
AUTH_PASSWORD=mia-password-sicura
```

### Connessione con Windows Authentication

Per usare l'autenticazione integrata di Windows (solo su host Windows), impostare `mode: integrated` nella voce server:

```yaml
servers:
  - name: production
    database:
      mode: integrated
      host: localhost
      name: MaintenanceDB
      # username e password vengono ignorati
```

### Connessione a una named instance

```yaml
servers:
  - name: production
    database:
      host: MYSERVER
      instance: SQLEXPRESS   # si connette a MYSERVER\SQLEXPRESS
      port: 0                # la porta viene ignorata quando si specifica l'istanza
```

### Connessione con TLS

```yaml
servers:
  - name: production
    database:
      encrypt: true
      trust_server_certificate: false  # false = verifica il certificato (raccomandato in produzione)
```

---

## Avvio

```bash
./maintenance-dashboard
```

Il programma stampa l'indirizzo di ascolto all'avvio:

```
2024/01/15 10:00:00 starting maintenance-dashboard on 0.0.0.0:8080
```

Apri il browser su `http://<host>:<port>`. Se l'autenticazione è abilitata, il browser mostrerà una finestra di login HTTP Basic.

### Controllo di salute

L'endpoint `/api/meta/health` non richiede autenticazione e risponde `200 OK` se il server è operativo:

```bash
curl http://localhost:8080/api/meta/health
```

---

## Pagine disponibili

| Percorso | Pagina | Filtri principali |
|---|---|---|
| `/servers` | Servers | — |
| `/` | Dashboard | Server, database, date, tipo comando, solo errori |
| `/fragmented-indexes` | Top Fragmented Indexes | Server, database, date, top N |
| `/modified-statistics` | Most Modified Statistics | Server, database, date, top N |
| `/operations-per-batch` | Operations Per Batch | Server, database, date, granularità |
| `/maintenance-summary` | Maintenance Summary | Server, database, date, granularità |
| `/backup-overview` | Backup Overview | Server, database, date, granularità |
| `/maintenance-errors` | Maintenance Errors | Server, database, tipo comando, date, granularità |
| `/long-running-operations` | Long Running Operations | Server, database, tipo comando, date, granularità, durata minima |

---

## API REST

Tutti gli endpoint (eccetto `/api/meta/health`) richiedono HTTP Basic Auth se `auth.enabled: true`.

| Metodo | Endpoint | Parametri query |
|---|---|---|
| `GET` | `/api/meta/health` | — |
| `GET` | `/api/meta/servers` | — |
| `GET` | `/api/meta/server-status` | — |
| `GET` | `/api/meta/filters` | `server` |
| `GET` | `/api/dashboard/summary` | filtri standard |
| `GET` | `/api/statistics/most-modified` | filtri standard + `limit` (1–500, default 50) |
| `GET` | `/api/indexes/top-fragmented` | filtri standard + `limit` (1–500, default 50) |
| `GET` | `/api/maintenance/summary` | filtri standard + `granularity` (day/week/month) |
| `GET` | `/api/operations/per-batch` | filtri standard + `granularity` |
| `GET` | `/api/backup/report` | filtri standard + `granularity` |
| `GET` | `/api/errors/report` | filtri standard + `granularity` + `limit` (1–500, default 100) |
| `GET` | `/api/longrunning/report` | filtri standard + `granularity` + `minDuration` (secondi, default 300) + `limit` |

**Filtri standard** — parametri query comuni a tutti gli endpoint di dati:

| Parametro | Tipo | Descrizione |
|---|---|---|
| `server` | stringa | Nome del server di destinazione (come definito in `servers[].name`); default: primo server |
| `database` | stringa | Filtra per nome database |
| `dateFrom` | `YYYY-MM-DD` | Data di inizio (inclusa) |
| `dateTo` | `YYYY-MM-DD` | Data di fine (inclusa) |
| `commandType` | stringa | Filtra per tipo di comando (es. `ALTER_INDEX`, `UPDATE_STATISTICS`) |
| `schema` | stringa | Filtra per schema |
| `object` | stringa | Filtra per nome oggetto |
| `onlyErrors` | `true`/`false` | Mostra solo le operazioni in errore |

Tutte le risposte hanno il formato:

```json
{ "data": { ... } }
```

Gli errori restituiscono:

```json
{ "error": { "code": "...", "message": "..." } }
```
