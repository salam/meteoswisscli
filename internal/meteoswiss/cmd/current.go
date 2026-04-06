package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/i18n"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var currentLimit int

func init() {
	rootCmd.AddCommand(currentCmd)
	currentCmd.Flags().IntVar(&currentLimit, "limit", 5, "Number of nearby stations to show when using place name/coordinates")
}

var currentCmd = &cobra.Command{
	Use:   "current [location]",
	Short: "Current weather measurements",
	Long: `Show current measurements from automatic weather stations.

Accepts a station code (e.g. SMA), place name (e.g. Zürich), PLZ (e.g. 8001),
or coordinates (e.g. 47.37,8.55). Place names and coordinates show the nearest stations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)
		measurements, err := client.GetCurrentMeasurements("")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if len(args) == 0 {
			return showMeasurements("", measurements)
		}

		input := args[0]

		// Try resolving as station code, place name, or coordinates
		resolved, resolveErr := geo.ResolveStation(input, currentLimit)
		if resolveErr != nil {
			output.Error(resolveErr.Error())
			os.Exit(1)
		}

		// Build set of station codes to show
		stationCodes := make(map[string]bool)
		for _, r := range resolved {
			stationCodes[strings.ToUpper(r.Station.Code)] = true
		}

		var filtered []api.StationMeasurement
		for _, m := range measurements {
			if stationCodes[m.Station] {
				filtered = append(filtered, m)
			}
		}

		if len(filtered) == 0 {
			// Fallback: try direct station code match (some codes may not be in embedded data)
			upper := strings.ToUpper(input)
			for _, m := range measurements {
				if m.Station == upper {
					filtered = append(filtered, m)
				}
			}
		}

		if len(filtered) == 0 {
			output.Error(fmt.Sprintf("no measurements found for %q", input))
			os.Exit(1)
		}

		label := ""
		if len(resolved) > 0 && resolved[0].Distance > 0 {
			// Location-based lookup — show what we resolved to
			if loc, err := geo.SearchLocation(input); err == nil {
				label = fmt.Sprintf("%s %s", loc.Name, loc.Kanton)
			}
		}

		return showMeasurements(label, filtered)
	},
}

func showMeasurements(label string, measurements []api.StationMeasurement) error {
	if !output.IsInteractive() {
		result := map[string]any{
			"measurements": measurements,
			"source":       source.MeteoSwiss,
		}
		if label != "" {
			result["location"] = label
		}
		output.JSON(result)
		return nil
	}

	title := i18n.T("Current Measurements")
	if label != "" {
		title += " — " + label
	}
	output.Section(title)
	headers := []string{i18n.T("STATION"), i18n.T("NAME"), i18n.T("TEMP"), i18n.T("HUMIDITY"), i18n.T("WIND"), i18n.T("GUSTS"), i18n.T("PRESSURE"), i18n.T("RAIN")}
	var rows [][]string
	for _, m := range measurements {
		name := ""
		if s := geo.LookupStation(m.Station); s != nil {
			name = s.Name
		}
		rows = append(rows, []string{
			m.Station,
			name,
			fmtVal(m.Temperature, "°C"),
			fmtVal(m.Humidity, "%"),
			fmtVal(m.WindSpeed, " km/h"),
			fmtVal(m.GustPeak, " km/h"),
			fmtVal(m.PressureQNH, " hPa"),
			fmtVal(m.Rainfall, " mm"),
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func fmtVal(val, unit string) string {
	if val == "" {
		return "-"
	}
	return val + unit
}
