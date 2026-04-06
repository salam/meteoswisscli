package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(radarCmd)
}

var radarCmd = &cobra.Command{
	Use:   "radar [rain|cloud|satellite]",
	Short: "Weather radar and satellite images",
	Long:  "View rain radar, cloud radar, or satellite imagery. Opens in browser (interactive maps, no static image available).",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		radarType := api.RadarRain
		if len(args) > 0 {
			switch args[0] {
			case "rain":
				radarType = api.RadarRain
			case "cloud":
				radarType = api.RadarCloud
			case "satellite":
				radarType = api.RadarSatellite
			default:
				output.Error(fmt.Sprintf("unknown radar type %q — use rain, cloud, or satellite", args[0]))
				os.Exit(1)
			}
		}

		url := api.GetRadarBrowserURL(radarType)

		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"type":   string(radarType),
				"url":    url,
				"source": source.MeteoSwiss,
			})
			return nil
		}

		fmt.Printf("Opening %s radar in browser...\n", radarType)
		if err := output.OpenBrowser(url); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", url)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
