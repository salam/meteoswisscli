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

var (
	stationsSearch string
	stationsNear   string
	stationsLimit  int
)

func init() {
	rootCmd.AddCommand(stationsCmd)
	stationsCmd.Flags().StringVar(&stationsSearch, "search", "", "Search stations by code")
	stationsCmd.Flags().StringVar(&stationsNear, "near", "", "Find stations near a place name, PLZ, or lat,lon")
	stationsCmd.Flags().IntVar(&stationsLimit, "limit", 10, "Number of nearby stations to show (with --near)")
}

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "List weather measurement stations",
	Long:  "List MeteoSwiss automatic weather stations. Use --near to find stations close to a location.",
	Example: `  meteoswiss stations --near Bern
  meteoswiss stations --search ZRH`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --near mode: use embedded station data with coordinates
		if stationsNear != "" {
			return showNearbyMeteoStations()
		}

		// Default: show from live measurements
		client := api.NewClientWithCache(Lang, ResponseCache)
		measurements, err := client.GetCurrentMeasurements("")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		seen := make(map[string]bool)
		var stations []api.StationMeasurement
		for _, m := range measurements {
			if seen[m.Station] {
				continue
			}
			seen[m.Station] = true
			if stationsSearch != "" && !strings.Contains(strings.ToUpper(m.Station), strings.ToUpper(stationsSearch)) {
				continue
			}
			stations = append(stations, m)
		}

		if !output.IsInteractive() {
			codes := make([]string, len(stations))
			for i, s := range stations {
				codes[i] = s.Station
			}
			output.JSON(map[string]any{
				"stations": codes,
				"count":    len(codes),
				"source":   source.MeteoSwiss,
			})
			return nil
		}

		output.Section(fmt.Sprintf("%s (%d)", i18n.T("Stations"), len(stations)))
		headers := []string{i18n.T("CODE"), i18n.T("TEMP"), i18n.T("WIND"), i18n.T("RAIN")}
		var rows [][]string
		for _, s := range stations {
			rows = append(rows, []string{
				s.Station,
				fmtVal(s.Temperature, "°C"),
				fmtVal(s.WindSpeed, " km/h"),
				fmtVal(s.Rainfall, " mm"),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}

func showNearbyMeteoStations() error {
	resolved, err := geo.ResolveStation(stationsNear, stationsLimit)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	label := stationsNear
	if loc, err := geo.SearchLocation(stationsNear); err == nil {
		label = fmt.Sprintf("%s %s", loc.Name, loc.Kanton)
	}

	if !output.IsInteractive() {
		type result struct {
			Code      string  `json:"code"`
			Name      string  `json:"name"`
			Elevation int     `json:"elevation"`
			Canton    string  `json:"canton"`
			Distance  float64 `json:"distance_km"`
		}
		results := make([]result, len(resolved))
		for i, r := range resolved {
			results[i] = result{
				Code: r.Station.Code, Name: r.Station.Name,
				Elevation: r.Station.Elevation, Canton: r.Station.Canton,
				Distance: float64(int(r.Distance*10)) / 10,
			}
		}
		output.JSON(map[string]any{"near": label, "stations": results, "count": len(results), "source": source.MeteoSwiss})
		return nil
	}

	output.Section(fmt.Sprintf("%s near %s (%d)", i18n.T("Stations"), label, len(resolved)))
	headers := []string{i18n.T("CODE"), i18n.T("NAME"), i18n.T("ELEVATION"), i18n.T("CANTON"), i18n.T("DISTANCE")}
	var rows [][]string
	for _, r := range resolved {
		rows = append(rows, []string{
			r.Station.Code, r.Station.Name,
			fmt.Sprintf("%dm", r.Station.Elevation), r.Station.Canton,
			fmt.Sprintf("%.1f km", r.Distance),
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}
