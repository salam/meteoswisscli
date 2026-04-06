# Release Notes

## v0.1.1 (Sun, Apr 6 11:15)

* INCA nowcast precipitation forecast: `radar --list` now shows frames extending up to +6h into the future using MeteoSwiss INCA data [claude]
* ASCII wind map: `wind --ascii` renders wind direction arrows color-coded by speed on a map of Switzerland [claude]
* Radar combines HDF5 past + INCA future frames in list, ASCII, and interactive modes [claude]

## v0.1.0 -- Initial Release (2026-04-06)

First public release of meteoswisscli, providing two CLI tools for Swiss weather and avalanche data.

### MeteoSwiss CLI

- **Weather forecast** with 3-day and 8-day views for any Swiss location
- **Current conditions** from 50+ MeteoSwiss measurement stations (temperature, humidity, wind, pressure, sunshine, rainfall)
- **Precipitation radar** with ASCII art rendering from HDF5 open data, time series playback, Swiss border/lake overlays, and interactive frame scrolling
- **Wind measurements** with speed, gusts, and direction
- **Pollen data** from 16 live monitoring stations with nearest-station resolution
- **Prose weather bulletins** from MeteoSwiss forecasters (north/south/west regions, with outlook)
- **Precipitation forecast** with 10-day probability data
- **Natural hazards** and **cloud cover** maps via browser
- **Station listing** with search and nearest-station lookup

### WhiteRisk CLI

- **Avalanche bulletins** from SLF with danger ratings, avalanche problems, and snowpack analysis (CAAML format)
- **Snow measurements** from 207 IMIS stations and study plots (1/3/7-day history)
- **Snow maps** (new snow, depth, comparison) with ASCII rendering
- **Avalanche activity** reports and **snow profiles** via browser

### Shared Features

- **Flexible location input**: postal codes, place names, coordinates, station codes
- **Embedded location database**: 3190+ Swiss settlements, zero network lookups for resolution
- **Four languages**: de, fr, it, en with auto-detection from system locale
- **Multiple output modes**: colored tables, JSON, ASCII art, browser, image save
- **Favorites system** for saving frequently used locations
- **Cross-platform**: macOS, Linux, Windows (amd64 + arm64)
- **Pure Go HDF5 reader** for radar data (no Python/h5py dependency)

### Technical Details

- Go 1.25, minimal dependencies (cobra, go-native-netcdf, x/term)
- Single static binary per tool (~10MB)
- All data from public, unauthenticated APIs
- Secure config storage with proper file permissions

### Data Sources

- MeteoSwiss App API (forecasts)
- Swiss Federal Open Data (measurements, radar, pollen)
- SLF Measurement API (IMIS snow stations)
- SLF Avalanche Service (CAAML bulletins)
- WhiteRisk (snow maps)
