package cmd

import (
	"fmt"

	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var windLevel string

func init() {
	rootCmd.AddCommand(windCmd)
	windCmd.Flags().StringVar(&windLevel, "level", "10m", "Wind level: 10m, 2000m, gust")
}

var windCmd = &cobra.Command{
	Use:   "wind",
	Short: "Wind data and animations",
	RunE: func(cmd *cobra.Command, args []string) error {
		urlMap := map[string]string{
			"10m":   "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-wind-10m",
			"2000m": "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-upper-winds-2000m",
			"gust":  "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-gust-peaks-10m",
		}
		url, ok := urlMap[windLevel]
		if !ok {
			output.Error(fmt.Sprintf("unknown wind level %q — use 10m, 2000m, or gust", windLevel))
			return nil
		}
		if !output.IsInteractive() {
			output.JSON(map[string]string{"level": windLevel, "url": url, "source": source.MeteoSwiss})
			return nil
		}
		fmt.Printf("Opening wind (%s) in browser...\n", windLevel)
		if err := output.OpenBrowser(url); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", url)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
