# SwissCLI Design Spec

Two Go CLI tools for Swiss weather and avalanche data: `meteoswiss` and `whiterisk`.

## Architecture

Monorepo with shared packages. Two binaries built from `cmd/meteoswiss/` and `cmd/whiterisk/`.

```
swisscli/
├── cmd/
│   ├── meteoswiss/main.go
│   └── whiterisk/main.go
├── internal/
│   ├── meteoswiss/
│   │   ├── cmd/          # cobra commands (one file per command)
│   │   └── api/          # HTTP client + one file per API resource
│   └── whiterisk/
│       ├── cmd/
│       └── api/
├── pkg/
│   ├── output/           # table, JSON, ASCII art, browser, image save
│   ├── config/           # JSON config at ~/.config/{app}/config.json
│   ├── geo/              # PLZ lookup, name search, region matching
│   └── source/           # attribution strings
├── go.mod
├── Makefile
└── README.md
```

### Dependencies

- `github.com/spf13/cobra` — CLI framework
- `modernc.org/sqlite` — pure Go SQLite for location DB (no CGO)
- stdlib for everything else: `net/http`, `encoding/json`, `encoding/csv`, `text/tabwriter`, `image/png`, `image/jpeg`, `database/sql`

## MeteoSwiss CLI

### API Sources (all unauthenticated)

| Base URL | Purpose |
|----------|---------|
| `https://app-prod-ws.meteoswiss-app.ch` | Forecast by PLZ (v1/plzDetail) |
| `https://data.geo.admin.ch` | Open data: measurements CSV, STAC collections (radar, precip, pollen, wind) |
| `https://s3-eu-central-1.amazonaws.com/app-prod-static-fra.meteoswiss-app.ch/v1/` | Static: location SQLite DB |

### Commands

| Command | Description | API Endpoint |
|---------|-------------|-------------|
| `meteoswiss forecast <location>` | Today's forecast | `/v1/plzDetail?plz={6digit}` |
| `meteoswiss forecast <location> --week` | 8-day forecast table | Same (returns 8 days) |
| `meteoswiss current [station]` | Current measurements | `data.geo.admin.ch/.../VQHA80.csv` |
| `meteoswiss radar rain` | Rain radar | Browser/save/ASCII |
| `meteoswiss radar cloud` | Cloud/satellite | Browser/save/ASCII |
| `meteoswiss radar satellite` | Satellite HRV | Browser/save/ASCII |
| `meteoswiss hazards [--region]` | Natural hazard warnings | Web/API |
| `meteoswiss wind [location]` | Wind (10m, 2000m, gusts) | Open data measurements |
| `meteoswiss pollen [--type TYPE]` | Pollen forecast & report | Open data pollen |
| `meteoswiss precipitation <location>` | Precipitation probability | plzDetail graph data |
| `meteoswiss clouds` | Cloud cover | Web/open data |
| `meteoswiss stations [--search TEXT]` | List stations | VQHA80 station metadata |
| `meteoswiss favorites add\|remove\|list` | Saved locations | Local config |

### Global Flags

- `--json` — force JSON output
- `--no-color` — disable colors
- `--lang de|fr|it|en` — override language (auto-detected from `$LANG`, fallback `de`)

### Location Resolution

1. Numeric input → treat as PLZ, pad to 6 digits (`8001` → `800100`)
2. Text input → search bundled location DB (SQLite, cached at `~/.config/meteoswiss/locations.db`)
3. `lat,lon` format (e.g. `47.37,8.55`) → reverse lookup to nearest PLZ

### Radar/Visual Output Flags

- Default: open in browser
- `--save [path]` — download image to file
- `--ascii` — ASCII art in terminal (Unicode block elements: ░▒▓█)

### Request Headers

```
Accept: application/json
Accept-Encoding: gzip
Accept-Language: {de|fr|it|en}
User-Agent: SwissCLI/1.0
```

## WhiteRisk CLI

### API Sources (all unauthenticated)

| Base URL | Purpose |
|----------|---------|
| `https://aws.slf.ch` | Avalanche bulletin (CAAML JSON), documents (PDF/GIF), warning regions |
| `https://measurement-api.slf.ch` | IMIS + study plot stations and measurements |
| `https://whiterisk.ch` | Snow map teasers (PNG) |

### Commands

| Command | Description | API Endpoint |
|---------|-------------|-------------|
| `whiterisk bulletin [location]` | Avalanche danger rating & text | `aws.slf.ch/api/bulletin/caaml/{lang}/json` |
| `whiterisk bulletin [location] --pdf` | Download bulletin PDF | `aws.slf.ch/api/bulletin/document/full/{lang}` |
| `whiterisk snow new` | New snow map (24/48/72h) | Browser/save/ASCII |
| `whiterisk snow depth` | Snow depth map | Browser/save/ASCII |
| `whiterisk snow compare` | Comparative snow depth | Browser/save/ASCII |
| `whiterisk measurements [station]` | IMIS measurements | `measurement-api.slf.ch/public/api/imis/...` |
| `whiterisk measurements [station] --type study-plot` | Study plot measurements | `.../study-plot/...` |
| `whiterisk measurements [station] --period 1\|3\|7` | History (days) | `period_in_days` param |
| `whiterisk stations [--search TEXT]` | List stations | `measurement-api.slf.ch/public/api/imis/stations` |
| `whiterisk stations --type imis\|study-plot` | Filter by type | Same |
| `whiterisk avalanches [--region]` | Current reported avalanches | GeoJSON endpoint |
| `whiterisk profiles [--region]` | Snow profiles | Snow profile backend |
| `whiterisk favorites add\|remove\|list` | Saved locations/stations | Local config |

### Bulletin Location Resolution

1. 4-digit number → match by region ID (e.g. `7231`)
2. Text → fuzzy-match region name (e.g. `Davos`)
3. `lat,lon` → find nearest warning region using GeoJSON boundaries from `aws.slf.ch/api/warningregion/warnregionDefinition/findByDate/geojson`
4. No location → overview table of all regions with danger levels

### SLF Measurement API (OpenAPI 3.1)

**IMIS Stations:** 207 automated stations. Fields: `code`, `label`, `lon`, `lat`, `elevation`, `country_code`, `canton_code`, `type` (SNOW_FLAT, WIND, FLOWCAPT).

**IMIS Measurements:** 30-min intervals. Fields: `HS` (snow height cm), `TA_30MIN_MEAN` (air temp C), `RH_30MIN_MEAN` (humidity %), `TSS_30MIN_MEAN` (snow surface temp C), `VW_30MIN_MEAN` (wind m/s), `VW_30MIN_MAX` (gust m/s), `DW_30MIN_MEAN` (wind dir deg), `RSWR_30MIN_MEAN` (reflected radiation W/m2).

**Study Plot Stations:** Manual measurement sites. Fields: `code`, `label`, `lon`, `lat`, `elevation`, `country_code`, `canton_code`.

**Study Plot Measurements:** Daily. Fields: `HS` (snow height cm), `HN_1D` (new snow 24h cm), `HNW_1D` (water equivalent mm).

**Period parameter:** `1`, `3`, or `7` days.

## Shared Packages

### `pkg/output/`

- `IsInteractive() bool` — checks TTY + `--json` flag
- `Table(headers, rows)` — tabwriter, colored header row
- `JSON(v)` — pretty-printed JSON to stdout
- `Error(msg)` — stderr (TTY) or JSON `{"error":...}` (pipe)
- `Section(title)` — `\n--- title ---\n`
- `OpenBrowser(url)` — `open` (macOS) / `xdg-open` (Linux)
- `SaveImage(url, path)` — HTTP GET → file
- `ASCIIMap(imageURL, width)` — download image, convert to Unicode block art (░▒▓█)

### `pkg/config/`

```go
type Config struct {
    Lang      string     `json:"lang"`
    Favorites []Favorite `json:"favorites"`
}

type Favorite struct {
    Name    string  `json:"name"`
    PLZ     string  `json:"plz,omitempty"`
    Region  string  `json:"region,omitempty"`
    Station string  `json:"station,omitempty"`
    Lat     float64 `json:"lat,omitempty"`
    Lon     float64 `json:"lon,omitempty"`
}
```

- MeteoSwiss: `~/.config/meteoswiss/config.json`
- WhiteRisk: `~/.config/whiterisk/config.json`
- Dir permissions: `0700`, file permissions: `0600`
- Env overrides: `METEOSWISS_LANG`, `WHITERISK_LANG`

### `pkg/geo/`

- `ResolvePLZ(input) (plz6digit, error)` — numeric/text/lat,lon
- `ResolveRegion(input) (regionID, error)` — ID/name/lat,lon
- Location DB: SQLite from MeteoSwiss S3, cached `~/.config/meteoswiss/locations.db`
- Warning regions: GeoJSON from aws.slf.ch, cached `~/.config/whiterisk/regions.geojson` (24h TTL)

### `pkg/source/`

```go
const MeteoSwiss = "Quelle: MeteoSchweiz; Source: MétéoSuisse; Fonte: MeteoSvizzera; Source: MeteoSwiss"
const SLF = "Quelle: SLF/WSL; Source: SLF/WSL"
```

## Language

Auto-detected from `$LANG` / `$LC_ALL` → extract `de`, `fr`, `it`, `en`. Fallback: `de`. Override: `--lang` flag or `METEOSWISS_LANG` / `WHITERISK_LANG` env var.

## Output Pattern

Every command:
1. Check `IsInteractive()`
2. TTY → colored table with section headers, attribution as last line
3. Pipe/`--json` → JSON object with `"source"` key containing attribution string

## Error Handling

- `RunE` on all commands (return errors)
- `output.Error(msg)` for user-facing errors
- `fmt.Errorf("context: %w", err)` for wrapping
- No `log` package
- Network: `"could not reach API. Check your internet connection"`
- Bad location: `"location not found. Try a PLZ code (e.g. 8001) or place name"`

## Build

```makefile
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
```

Two binaries: `meteoswiss` and `whiterisk`. Cross-compilation for darwin/linux (amd64/arm64) + windows/amd64.

## Cache

- Location DB: downloaded once, refresh with `meteoswiss stations --update`
- Warning regions GeoJSON: 24h TTL
- Weather/measurement data: never cached (always fresh)

## Verified Endpoints

| Endpoint | Status |
|----------|--------|
| `app-prod-ws.meteoswiss-app.ch/v1/plzDetail?plz=800100` | 200 OK |
| `data.geo.admin.ch/ch.meteoschweiz.messwerte-aktuell/VQHA80.csv` | 200 OK |
| `data.geo.admin.ch/api/stac/v1/collections/ch.meteoschweiz.ogd-local-forecasting` | 200 OK |
| `data.geo.admin.ch/api/stac/v1/collections/ch.meteoschweiz.ogd-radar-precip` | 200 OK |
| `measurement-api.slf.ch/public/api/imis/stations` | 200 OK (207 stations) |
| `aws.slf.ch/api/bulletin/caaml/en/json` | 200 OK (6 bulletins) |
