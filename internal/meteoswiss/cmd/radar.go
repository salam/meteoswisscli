package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	radarSave  string
	radarASCII bool
	radarWidth int
)

func init() {
	rootCmd.AddCommand(radarCmd)
	radarCmd.Flags().StringVar(&radarSave, "save", "", "Save image to file path")
	radarCmd.Flags().BoolVar(&radarASCII, "ascii", false, "Render as ASCII art in terminal")
	radarCmd.Flags().IntVar(&radarWidth, "width", 80, "ASCII art width in columns")
}

var radarCmd = &cobra.Command{
	Use:   "radar [rain|cloud|satellite]",
	Short: "Weather radar and satellite images",
	Long:  "View rain radar, cloud radar, or satellite imagery. Default: opens in browser.",
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

		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"type":        string(radarType),
				"browser_url": api.GetRadarBrowserURL(radarType),
				"image_url":   api.GetRadarImageURL(radarType),
				"source":      source.MeteoSwiss,
			})
			return nil
		}

		if radarASCII {
			fmt.Printf("Fetching %s radar...\n", radarType)
			if err := output.ASCIIMap(api.GetRadarImageURL(radarType), radarWidth); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("\n%s\n", source.MeteoSwiss)
			return nil
		}

		if radarSave != "" {
			path := radarSave
			if filepath.Ext(path) == "" {
				path += ".png"
			}
			fmt.Printf("Saving %s radar to %s...\n", radarType, path)
			if err := output.SaveImage(api.GetRadarImageURL(radarType), path); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("Saved to %s\n", path)
			fmt.Printf("\n%s\n", source.MeteoSwiss)
			return nil
		}

		url := api.GetRadarBrowserURL(radarType)
		fmt.Printf("Opening %s radar in browser...\n", radarType)
		if err := output.OpenBrowser(url); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", url)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
