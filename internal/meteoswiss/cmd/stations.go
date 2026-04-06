package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var stationsSearch string

func init() {
	rootCmd.AddCommand(stationsCmd)
	stationsCmd.Flags().StringVar(&stationsSearch, "search", "", "Search stations by code")
}

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "List weather measurement stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)
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

		output.Section(fmt.Sprintf("Stations (%d)", len(stations)))
		headers := []string{"CODE", "TEMP", "WIND", "RAIN"}
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
