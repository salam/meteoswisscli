package cmd

import (
	"fmt"

	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

const avalanchesURL = "https://whiterisk.ch/de/conditions/current-avalanches"

func init() {
	rootCmd.AddCommand(avalanchesCmd)
}

var avalanchesCmd = &cobra.Command{
	Use:   "avalanches",
	Short: "Current reported avalanches",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{"url": avalanchesURL, "source": source.SLF})
			return nil
		}
		fmt.Println("Opening current avalanches in browser...")
		if err := output.OpenBrowser(avalanchesURL); err != nil {
			fmt.Printf("Visit: %s\n", avalanchesURL)
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	},
}
