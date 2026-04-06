package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init() {
	rootCmd.AddCommand(manCmd)
}

var manCmd = &cobra.Command{
	Use:    "man <output-dir>",
	Short:  "Generate man pages",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		header := &doc.GenManHeader{
			Title:   "WHITERISK",
			Section: "1",
			Source:  "whiterisk " + version,
		}
		if err := doc.GenManTree(cmd.Root(), header, args[0]); err != nil {
			return err
		}
		fmt.Printf("Man pages generated in %s\n", args[0])
		return nil
	},
}
