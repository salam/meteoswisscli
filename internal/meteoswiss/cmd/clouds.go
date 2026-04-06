package cmd

import (
	"fmt"

	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

const cloudsURL = "https://www.meteoswiss.admin.ch/services-and-publications/applications/cloud-cover.html#tab=cloud-coverage"

func init() {
	rootCmd.AddCommand(cloudsCmd)
}

var cloudsCmd = &cobra.Command{
	Use:     "clouds",
	Short:   "Cloud cover visualization",
	Example: `  meteoswiss clouds`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{"url": cloudsURL, "source": source.MeteoSwiss})
			return nil
		}
		fmt.Println("Opening cloud cover in browser...")
		if err := output.OpenBrowser(cloudsURL); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", cloudsURL)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
