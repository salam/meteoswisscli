# meteoswisscli

Two command-line tools for Swiss weather and avalanche data, written in Go.

- **`meteoswiss`** -- Weather forecasts, current conditions, radar, wind, pollen, and more from [MeteoSwiss](https://www.meteoswiss.admin.ch/)
- **`whiterisk`** -- Avalanche bulletins, snow measurements, and snow maps from [SLF/WSL](https://www.slf.ch/) and [WhiteRisk](https://whiterisk.ch/)

All data comes from public, unauthenticated APIs and Swiss open data.

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap salam/tap
brew install meteoswisscli
```

This installs both `meteoswiss` and `whiterisk` binaries.

### Download binary

Grab the latest release for your platform from [GitHub Releases](https://github.com/salam/meteoswisscli/releases) and add it to your `PATH`.

### From source

Requires Go 1.25+.

```bash
git clone https://github.com/salam/meteoswisscli.git
cd meteoswisscli
make build
```

This produces two binaries: `meteoswiss` and `whiterisk`.

### Cross-compile for all platforms

```bash
make all
```

Builds for macOS (amd64/arm64), Linux (amd64/arm64), and Windows (amd64) into `dist/`.

### Shell completions

```bash
# Bash
meteoswiss completion bash > /etc/bash_completion.d/meteoswiss

# Zsh
meteoswiss completion zsh > "${fpath[1]}/_meteoswiss"

# Fish
meteoswiss completion fish > ~/.config/fish/completions/meteoswiss.fish
```

Same for `whiterisk`.

## Demo

### Forecast

```
$ meteoswiss forecast Zürich

--- 800100 Zürich ZH ---

--- Aktuelles Wetter ---
  2025-04-06 12:00  18.3°C  (Icon: 2)

--- Vorhersage ---
DATUM        SYMBOL         MIN    MAX    NIEDERSCHLAG
2025-04-06   Mostly sunny   9°C    19°C   0.0 mm
2025-04-07   Partly cloudy  8°C    17°C   0.2 mm
2025-04-08   Rain           6°C    12°C   8.5 mm
```

### Interactive Radar

```
$ meteoswiss radar rain --interactive --frames 12

Precipitation Radar — 2025-04-06 11:50  [12/12]
← prev  → next  q quit

                    ·····
              ······       ····
           ···    ~~~~~        ···
         ··     ~~~~~~~~  ▒▒▓    ··
        ·    ~~~~~~         ▒▒▓▓   ·
       ·   ~~~~              ░▒▒▓   ·
      ·                       ░░▒    ·
       ·                             ·
        ·   ~~~~                   ··
         ··  ~~  ~~~~~~          ··
           ···       ~~~~~    ···
              ······       ····
                    ·····

Legend: ░ <0.1mm  ▒ <0.5mm  ▓ <1mm  · border  ~ lake
```

### Avalanche Bulletin

```
$ whiterisk bulletin Davos

--- Lawinenbulletin ---
  Regions: Davos (7231), Flüela (7232)
  Valid:   2025-04-06T08:00:00+02:00 → 2025-04-07T08:00:00+02:00
  Danger:  2 — Moderate (above 2200m)
  Problems:
    - wind_slab — N/NE/E above 2200m
```

## Quick Start

```bash
# Weather forecast for Zurich
meteoswiss forecast Zürich

# 8-day forecast by postal code
meteoswiss forecast 8001 --week

# Current measurements from nearby stations
meteoswiss current Bern

# Precipitation radar as ASCII art in the terminal
meteoswiss radar rain --ascii

# Wind measurements near a location
meteoswiss wind "Arosa GR"

# Pollen data
meteoswiss pollen Basel

# Prose weather bulletin
meteoswiss bulletin --region north

# Avalanche bulletin for Davos
whiterisk bulletin Davos

# Snow measurements from an IMIS station (7-day history)
whiterisk measurements DAV2 --period 7

# Snow depth map as ASCII art
whiterisk snow depth --ascii
```

## Location Input

Both tools accept flexible location inputs:

| Format | Example | Description |
|--------|---------|-------------|
| Place name | `Zürich`, `Bern`, `Arosa GR` | Searches 3190+ Swiss settlements |
| Postal code | `8001`, `3000` | Auto-padded to 6-digit MeteoSwiss PLZ |
| Coordinates | `47.37,8.55` | Latitude,longitude (WGS84) |
| Station code | `SMA`, `ZRH` | Direct station lookup |

## Output Modes

| Mode | When | Description |
|------|------|-------------|
| **Table** | Interactive terminal (TTY) | Colored, formatted tables with section headers |
| **JSON** | Piped output or `--json` | Machine-readable, includes `"source"` attribution |
| **ASCII art** | `--ascii` flag (radar, snow) | Unicode block rendering in terminal |
| **Browser** | Default for maps/hazards | Opens system browser |
| **Save** | `--save` flag | Downloads image to disk |

## Global Flags

```
--json        Force JSON output
--no-color    Disable ANSI colors
--lang        Override language: de, fr, it, en (auto-detected from $LANG)
--help, -h    Show help
--version     Show version
```

## Commands

### meteoswiss

| Command | Description |
|---------|-------------|
| `forecast <location>` | 3-day forecast (or 8-day with `--week`) |
| `current [location]` | Current measurements from nearby stations |
| `wind [location]` | Wind speed, gusts, and direction |
| `radar {rain\|cloud\|satellite}` | Precipitation/cloud/satellite imagery |
| `precipitation <location>` | 10-day precipitation forecast |
| `pollen [station]` | Pollen concentrations from 16 stations |
| `bulletin` | Prose weather forecast text |
| `hazards` | Natural hazard warning map |
| `clouds` | Cloud cover map |
| `stations` | List measurement stations |
| `favorites` | Manage saved locations |

### whiterisk

| Command | Description |
|---------|-------------|
| `bulletin [region]` | Avalanche danger ratings (CAAML) |
| `measurements <station>` | Snow depth, temperature, wind (IMIS/study-plot) |
| `stations` | List snow measurement stations |
| `snow {new\|depth\|compare}` | Snow maps |
| `avalanches` | Current avalanche activity reports |
| `profiles` | Snow profile data |
| `favorites` | Manage saved stations/regions |

## Configuration

Both tools store favorites and preferences in `~/.config/{app}/config.json`:

```bash
# Save a favorite location
meteoswiss favorites add "Home" --plz 8001
whiterisk favorites add "Davos" --region 7231

# List favorites
meteoswiss favorites list
```

Language is auto-detected from your system locale (`$LANG`) with fallback to German. Override with `--lang` or the `METEOSWISS_LANG` / `WHITERISK_LANG` environment variables.

## Data Sources

| Source | Data | API |
|--------|------|-----|
| [MeteoSwiss App API](https://www.meteoswiss.admin.ch/) | Forecasts, warnings | `app-prod-ws.meteoswiss-app.ch` |
| [Swiss Open Data](https://data.geo.admin.ch/) | Measurements, radar, pollen | `data.geo.admin.ch` (STAC + direct) |
| [SLF Measurement API](https://www.slf.ch/) | IMIS stations, snow data | `measurement-api.slf.ch` |
| [SLF Avalanche Service](https://www.slf.ch/) | Avalanche bulletins (CAAML) | `aws.slf.ch` |
| [WhiteRisk](https://whiterisk.ch/) | Snow maps | `whiterisk.ch` |

All APIs are public and require no authentication.

## Architecture

Monorepo with two binaries sharing common packages:

```
cmd/meteoswiss/          Entry point
cmd/whiterisk/           Entry point
internal/meteoswiss/     Commands + API client
internal/whiterisk/      Commands + API client
pkg/output/              Table, JSON, ASCII art, browser, image rendering
pkg/config/              Config file handling, favorites
pkg/geo/                 Location search, PLZ lookup, station resolution
pkg/i18n/                Translations (de/fr/it/en)
pkg/source/              Attribution strings
```

**Dependencies:** `cobra` for CLI framework, `go-native-netcdf` for HDF5 radar data, `x/term` for terminal detection. Everything else is Go stdlib.

## License

MIT
