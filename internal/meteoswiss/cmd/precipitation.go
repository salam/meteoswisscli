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

func init() {
	rootCmd.AddCommand(precipitationCmd)
}

var precipitationCmd = &cobra.Command{
	Use:   "precipitation <location>",
	Short: "Precipitation probability",
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
			output.JSON(map[string]any{"forecast": detail.Forecast, "graph": detail.Graph, "source": source.MeteoSwiss})
			return nil
		}
		output.Section("Precipitation Forecast")
		headers := []string{"DATE", "PRECIP", "MIN", "MAX"}
		var rows [][]string
		for _, d := range detail.Forecast {
			rows = append(rows, []string{
				d.DayDate,
				fmt.Sprintf("%.1f mm", d.Precipitation),
				fmt.Sprintf("%.1f mm", d.PrecipitationMin),
				fmt.Sprintf("%.1f mm", d.PrecipitationMax),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
