package cmd

import (
	"fmt"

	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

const profilesURL = "https://whiterisk.ch/de/conditions/snow-profiles"

func init() {
	rootCmd.AddCommand(profilesCmd)
}

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Short:   "Snow profiles",
	Example: `  whiterisk profiles`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{"url": profilesURL, "source": source.SLF})
			return nil
		}
		fmt.Println("Opening snow profiles in browser...")
		if err := output.OpenBrowser(profilesURL); err != nil {
			fmt.Printf("Visit: %s\n", profilesURL)
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	},
}
