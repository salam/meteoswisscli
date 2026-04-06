package cmd

import (
	"fmt"

	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var pollenType string

func init() {
	rootCmd.AddCommand(pollenCmd)
	pollenCmd.Flags().StringVar(&pollenType, "type", "all", "Pollen type filter")
}

var pollenCmd = &cobra.Command{
	Use:   "pollen",
	Short: "Pollen forecast and report",
	RunE: func(cmd *cobra.Command, args []string) error {
		mapURL := fmt.Sprintf("https://www.meteoswiss.admin.ch/services-and-publications/applications/pollen-forecast.html#tab=pollen-map&pollen=%s", pollenType)
		forecastURL := fmt.Sprintf("https://www.meteoswiss.admin.ch/services-and-publications/applications/pollen-forecast.html#tab=pollen-forecast&pollen=%s", pollenType)
		if !output.IsInteractive() {
			output.JSON(map[string]string{"map_url": mapURL, "forecast_url": forecastURL, "pollen_type": pollenType, "source": source.MeteoSwiss})
			return nil
		}
		fmt.Println("Opening pollen forecast in browser...")
		if err := output.OpenBrowser(forecastURL); err != nil {
			fmt.Printf("Could not open browser. Visit:\n  Map: %s\n  Forecast: %s\n", mapURL, forecastURL)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
