package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/i18n"
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
	Example: `  meteoswiss precipitation Basel`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		location, err := getLocationArg(args, "meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		resolved, err := geo.ResolvePLZ(location)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		printCoordinateResolution(location, resolved)
		client := api.NewClientWithCache(Lang, ResponseCache)
		detail, err := client.GetForecast(resolved.PLZ)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}
		if !output.IsInteractive() {
			result := map[string]any{"forecast": detail.Forecast, "graph": detail.Graph, "source": source.MeteoSwiss}
			if resolved.Location != nil {
				result["location"] = resolved.Label()
			}
			output.JSON(result)
			return nil
		}
		title := i18n.T("Precipitation Forecast")
		if resolved.Location != nil {
			title += " — " + resolved.Label()
		}
		output.Section(title)
		headers := []string{i18n.T("DATE"), i18n.T("PRECIP"), i18n.T("MIN"), i18n.T("MAX")}
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
