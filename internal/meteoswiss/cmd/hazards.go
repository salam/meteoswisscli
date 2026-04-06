package cmd

import (
	"fmt"

	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

const hazardsURL = "https://www.meteoswiss.admin.ch/services-and-publications/applications/hazards.html#tab=natural-hazards-map"

func init() {
	rootCmd.AddCommand(hazardsCmd)
}

var hazardsCmd = &cobra.Command{
	Use:   "hazards",
	Short: "Natural hazard warnings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{"url": hazardsURL, "source": source.MeteoSwiss})
			return nil
		}
		fmt.Println("Opening natural hazards map in browser...")
		if err := output.OpenBrowser(hazardsURL); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", hazardsURL)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
