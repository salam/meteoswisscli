package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var weekFlag bool

func init() {
	rootCmd.AddCommand(forecastCmd)
	forecastCmd.Flags().BoolVar(&weekFlag, "week", false, "Show 8-day forecast")
}

var forecastCmd = &cobra.Command{
	Use:   "forecast <location>",
	Short: "Weather forecast for a location",
	Long:  "Show weather forecast by PLZ code (e.g. 8001), place name, or lat,lon coordinates.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		plz, err := geo.ParsePLZ(args[0])
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		client := api.NewClient(Lang)
		detail, err := client.GetForecast(plz)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"currentWeather": detail.CurrentWeather,
				"forecast":       detail.Forecast,
				"warnings":       detail.Warnings,
				"source":         source.MeteoSwiss,
			})
			return nil
		}

		output.Section("Current Weather")
		cw := detail.CurrentWeather
		fmt.Printf("  %s  %.1f°C  (Icon: %d)\n", cw.TimeFormatted(), cw.Temperature, cw.Icon)

		if weekFlag {
			output.Section("8-Day Forecast")
		} else {
			output.Section("Forecast")
		}

		days := detail.Forecast
		if !weekFlag && len(days) > 3 {
			days = days[:3]
		}

		headers := []string{"DATE", "ICON", "MIN", "MAX", "PRECIP"}
		var rows [][]string
		for _, d := range days {
			rows = append(rows, []string{
				d.DayDate,
				fmt.Sprintf("%d", d.IconDay),
				fmt.Sprintf("%.0f°C", d.TemperatureMin),
				fmt.Sprintf("%.0f°C", d.TemperatureMax),
				fmt.Sprintf("%.1f mm", d.Precipitation),
			})
		}
		output.Table(headers, rows)

		if len(detail.Warnings) > 0 {
			output.Section("Warnings")
			for _, w := range detail.Warnings {
				fmt.Printf("  [Level %d] %s (%s — %s)\n", w.Level, w.Text, w.ValidFrom, w.ValidTo)
			}
		}

		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
