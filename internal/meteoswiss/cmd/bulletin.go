package cmd

import (
	"fmt"

	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var bulletinRegion string

func init() {
	rootCmd.AddCommand(bulletinCmd)
	bulletinCmd.Flags().StringVar(&bulletinRegion, "region", "", "Region: north, south, west (default: overview)")
}

var bulletinCmd = &cobra.Command{
	Use:   "bulletin",
	Short: "Weather forecast bulletin (prose)",
	Long: `Open the MeteoSwiss weather forecast bulletin in the browser.

The bulletin contains prose weather forecasts for Switzerland:
  --region north   Northern Switzerland
  --region south   Southern Switzerland (Ticino, Engadin)
  --region west    Western Switzerland
  Without --region: overview for all of Switzerland`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlMap := map[string]string{
			"":      "https://www.meteoswiss.admin.ch/weather/weather-and-climate-from-a-to-z/weather-reports.html",
			"north": "https://www.meteoswiss.admin.ch/weather/weather-and-climate-from-a-to-z/weather-reports.html#tab=weather-report-north",
			"south": "https://www.meteoswiss.admin.ch/weather/weather-and-climate-from-a-to-z/weather-reports.html#tab=weather-report-south",
			"west":  "https://www.meteoswiss.admin.ch/weather/weather-and-climate-from-a-to-z/weather-reports.html#tab=weather-report-west",
		}

		url, ok := urlMap[bulletinRegion]
		if !ok {
			output.Error(fmt.Sprintf("unknown region %q — use north, south, or west", bulletinRegion))
			return nil
		}

		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"region": bulletinRegion,
				"url":    url,
				"source": source.MeteoSwiss,
			})
			return nil
		}

		region := "Switzerland"
		if bulletinRegion != "" {
			region = bulletinRegion
		}
		fmt.Printf("Opening weather bulletin (%s) in browser...\n", region)
		if err := output.OpenBrowser(url); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", url)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
