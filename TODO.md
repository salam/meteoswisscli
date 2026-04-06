# TODO

## Release Prep

- [x] Add LICENSE file (MIT)
- [ ] Tag v0.1.0
- [x] Set up goreleaser for automated release builds
- [x] Add install instructions for Homebrew tap
- [x] Generate and register shell completions (bash, zsh, fish -- cobra supports this)
- [x] Generate man pages from cobra commands

## Features

- [x] Satellite ASCII rendering (similar to radar, for cloud/satellite imagery)
- [x] HTTP response caching (measurements, bulletins) with configurable TTL
- [x] Parallel API calls for pollen stations
- [x] `--no-border` / `--no-lakes` flags for radar ASCII overlay control
- [x] Translate pollen species names (currently English only)
- [x] Reverse geocoding for coordinate inputs (show resolved place name)
- [x] Default location from config/favorite (skip location arg for forecast/current)

## Distribution

- [x] Homebrew formula / tap
- [x] GitHub Actions CI/CD pipeline (build, test, release)
- [ ] AUR package
- [ ] Snap / Flatpak
- [ ] Docker image

## Documentation

- [x] Add usage GIFs/screenshots to README (include the interactive radar sequence)
- [x] Man pages
- [x] `--help` examples for each command

## Testing

- [x] Integration tests for command output (table and JSON modes)
- [x] Test coverage for all commands (currently unit tests only for packages)
- [x] E2E end-to-end testing
- [x] CI test pipeline
