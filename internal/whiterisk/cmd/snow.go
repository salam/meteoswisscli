package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/salam/swissmeteocli/internal/whiterisk/api"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	snowSave  string
	snowASCII bool
	snowWidth int
)

func init() {
	rootCmd.AddCommand(snowCmd)
	snowCmd.AddCommand(snowNewCmd)
	snowCmd.AddCommand(snowDepthCmd)
	snowCmd.AddCommand(snowCompareCmd)
	for _, c := range []*cobra.Command{snowNewCmd, snowDepthCmd, snowCompareCmd} {
		c.Flags().StringVar(&snowSave, "save", "", "Save image to file")
		c.Flags().BoolVar(&snowASCII, "ascii", false, "ASCII art mode")
		c.Flags().IntVar(&snowWidth, "width", 80, "ASCII art width")
	}
}

var snowCmd = &cobra.Command{
	Use:   "snow",
	Short: "Snow maps and data",
	Example: `  whiterisk snow depth --ascii
  whiterisk snow new --save newsnow.png`,
}

var snowNewCmd = &cobra.Command{
	Use:   "new",
	Short: "New snow map",
	RunE:  snowMapRunner(api.SnowMapNew),
}

var snowDepthCmd = &cobra.Command{
	Use:   "depth",
	Short: "Snow depth map",
	RunE:  snowMapRunner(api.SnowMapDepth),
}

var snowCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Comparative snow depth map",
	RunE:  snowMapRunner(api.SnowMapCompare),
}

func snowMapRunner(mapType api.SnowMapType) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		browserURL := api.GetSnowBrowserURL(mapType)
		imageURL := api.GetSnowTeaserURL(mapType)

		if !output.IsInteractive() {
			output.JSON(map[string]string{"type": string(mapType), "browser_url": browserURL, "image_url": imageURL, "source": source.SLF})
			return nil
		}

		if snowASCII {
			fmt.Printf("Fetching %s snow map...\n", mapType)
			if err := output.ASCIIMap(imageURL, snowWidth); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("\n%s\n", source.SLF)
			return nil
		}

		if snowSave != "" {
			path := snowSave
			if filepath.Ext(path) == "" {
				path += ".png"
			}
			fmt.Printf("Saving %s snow map to %s...\n", mapType, path)
			if err := output.SaveImage(imageURL, path); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("Saved to %s\n", path)
			fmt.Printf("\n%s\n", source.SLF)
			return nil
		}

		fmt.Printf("Opening %s snow map in browser...\n", mapType)
		if err := output.OpenBrowser(browserURL); err != nil {
			fmt.Printf("Visit: %s\n", browserURL)
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	}
}
