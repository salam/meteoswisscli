package cmd

import (
	"fmt"
	"strings"

	"github.com/salam/swissmeteocli/pkg/config"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/output"
)

// getLocationArg returns the location from args or falls back to the default location in config.
func getLocationArg(args []string, appName string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	cfg, err := config.LoadDefault(appName)
	if err == nil && cfg.DefaultLocation != "" {
		return cfg.DefaultLocation, nil
	}
	return "", fmt.Errorf("no location specified. Use a place name, PLZ, or set a default with `%s favorites default <location>`", appName)
}

// printCoordinateResolution prints a line showing what coordinates resolved to, if the input looks like coordinates.
func printCoordinateResolution(input string, resolved *geo.ResolvedLocation) {
	if !output.IsInteractive() {
		return
	}
	if !strings.Contains(input, ",") {
		return
	}
	if resolved == nil || resolved.Location == nil {
		return
	}
	fmt.Printf("Resolved: %s → %s %s (PLZ %s)\n", input, resolved.Location.Name, resolved.Location.Kanton, resolved.Location.PLZ)
}
