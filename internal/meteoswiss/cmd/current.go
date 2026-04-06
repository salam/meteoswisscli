package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/i18n"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(currentCmd)
}

var currentCmd = &cobra.Command{
	Use:   "current [station]",
	Short: "Current weather measurements",
	Long:  "Show current measurements from automatic weather stations. Optionally filter by station code (e.g. SMA for Zürich).",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)
		measurements, err := client.GetCurrentMeasurements("")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if len(args) > 0 {
			search := strings.ToUpper(args[0])
			var filtered []api.StationMeasurement
			for _, m := range measurements {
				if m.Station == search {
					filtered = append(filtered, m)
				}
			}
			if len(filtered) == 0 {
				output.Error(fmt.Sprintf("station %q not found", args[0]))
				os.Exit(1)
			}
			measurements = filtered
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"measurements": measurements,
				"source":       source.MeteoSwiss,
			})
			return nil
		}

		output.Section(i18n.T("Current Measurements"))
		headers := []string{i18n.T("STATION"), i18n.T("TEMP"), i18n.T("HUMIDITY"), i18n.T("WIND"), i18n.T("GUSTS"), i18n.T("PRESSURE"), i18n.T("RAIN")}
		var rows [][]string
		for _, m := range measurements {
			rows = append(rows, []string{
				m.Station,
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
	},
}

func fmtVal(val, unit string) string {
	if val == "" {
		return "-"
	}
	return val + unit
}
