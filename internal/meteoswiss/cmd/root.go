package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/pkg/config"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	langFlag string
)

var rootCmd = &cobra.Command{
	Use:     "meteoswiss",
	Short:   "CLI for MeteoSwiss weather data",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		Lang = config.DetectLangWithEnv(langFlag, "METEOSWISS_LANG")
	},
}

var Lang string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&output.ForceJSON, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&output.NoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "Language (de, fr, it, en)")
}
