# Features

## MeteoSwiss CLI

### Weather Forecast
- 3-day and 8-day forecasts with min/max temperature, precipitation, and weather icons
- Weather icon descriptions with legend
- Active warnings (type, severity, valid time)
- Supports PLZ codes, place names, coordinates, and favorites

### Current Conditions
- Real-time measurements from 50+ MeteoSwiss weather stations
- Temperature, humidity, wind, pressure, sunshine, rainfall
- Auto-resolves nearest stations from place name or coordinates
- Configurable number of stations (`--limit`)

### Wind
- Wind speed, gust peaks, and direction from measurement stations
- Opens MeteoSwiss wind animation in browser (`--browser`)

### Precipitation Radar
- Three modes: rain, cloud, satellite
- ASCII art rendering in terminal with Unicode block characters and color intensity legend
- HDF5 open data: 640x710 grid at 1km resolution (CombiPrecip)
- Time series playback with `--frames` (10-minute intervals, up to 144 frames/24h)
- Interactive mode: scroll through frames in terminal
- Swiss border and lake outlines overlay
- Save images to disk or open in browser
- Configurable width (`--width`)

### Pollen
- Concentrations from 16 live pollen monitoring stations across Switzerland
- Species: Alder, Ash, Birch, Grasses, Hazel, Oak
- Auto-finds nearest station from place name, PLZ, or coordinates
- Multi-day view (`--days`)

### Weather Bulletin
- Prose weather forecast text (human-written by MeteoSwiss forecasters)
- Three regions: north, south, west
- Extended outlook with `--outlook`
- Available in de/fr/it

### Precipitation Forecast
- 10-day precipitation probability graph data
- Min, max, and average values

### Natural Hazards
- Opens the MeteoSwiss natural hazard warning map in browser

### Cloud Cover
- Opens cloud cover visualization in browser

### Stations
- List all MeteoSwiss measurement stations with metadata
- Search by station code (`--search`)
- Find nearest stations from place name or coordinates (`--near`)
- Current readings shown alongside station info

### Favorites
- Save, remove, and list favorite locations
- Stored in `~/.config/meteoswiss/config.json`

---

## WhiteRisk CLI

### Avalanche Bulletin
- Official SLF avalanche danger ratings from CAAML data
- Danger levels by elevation band (above/below treeline)
- Avalanche problems: type, aspects, elevation ranges
- Snowpack structure and weather forecast comments
- Travel advisory text
- Download full PDF bulletin (`--pdf`)

### Snow Measurements
- IMIS (automated) and study-plot (manual) station data
- 30-minute intervals (IMIS): snow depth, temperature, humidity, wind
- Daily aggregations (study-plot): snow depth, new snow, water equivalent
- History periods: 1, 3, or 7 days (`--period`)

### Snow Stations
- 207 IMIS automated snow measurement stations
- Study-plot manual observation sites
- Search by name/code, find nearest from coordinates

### Snow Maps
- Three map types: new snow, snow depth, comparison
- ASCII art rendering in terminal
- Save to disk or open in browser

### Avalanche Activity
- Current reported avalanche observations (opens in browser)

### Snow Profiles
- Snow profile measurement data (opens in browser)

### Favorites
- Save, remove, and list favorite stations and regions
- Stored in `~/.config/whiterisk/config.json`

---

## Shared Capabilities

### Location Resolution
- 3190+ Swiss settlements embedded in binary (no network lookup needed)
- Postal code input with auto-padding (e.g. `8001` -> `800100`)
- Place name search with canton disambiguation (e.g. `Arosa GR`)
- Municipality and settlement name matching
- Coordinate input (`lat,lon`) with nearest-location reverse lookup
- Haversine distance calculation for station proximity

### Internationalization
- Four languages: German, French, Italian, English
- Auto-detection from system locale (`$LANG`)
- Override via `--lang` flag or environment variable
- 40+ translated UI strings (section titles, table headers, labels)

### Output Formats
- Interactive terminal: colored tables with `tabwriter`, section headers, attribution footer
- JSON: pretty-printed with source attribution for piping and scripting
- ASCII art: Unicode block elements (░▒▓█) for radar and snow maps
- Browser: cross-platform launch (macOS/Linux/Windows)
- Image save: PNG/JPEG download to disk

### Configuration
- Per-app config at `~/.config/{app}/config.json`
- Secure file permissions (0700 directory, 0600 file)
- Favorites with name, PLZ, station code, region, coordinates

### Cross-Platform
- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)
- Single static binary, no runtime dependencies
