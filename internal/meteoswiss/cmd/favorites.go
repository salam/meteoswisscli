package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/pkg/config"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(favoritesCmd)
	favoritesCmd.AddCommand(favoritesAddCmd)
	favoritesCmd.AddCommand(favoritesRemoveCmd)
	favoritesCmd.AddCommand(favoritesListCmd)
	favoritesCmd.AddCommand(favoritesDefaultCmd)
}

var favoritesCmd = &cobra.Command{
	Use:   "favorites",
	Short: "Manage saved locations",
	Example: `  meteoswiss favorites list
  meteoswiss favorites add Home 8001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return favoritesListCmd.RunE(cmd, args)
	},
}

var favoritesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if !output.IsInteractive() {
			output.JSON(cfg.Favorites)
			return nil
		}

		if len(cfg.Favorites) == 0 {
			fmt.Println("No favorites saved. Use `meteoswiss favorites add <name> <plz>` to add one.")
			return nil
		}

		headers := []string{"NAME", "PLZ", "STATION"}
		var rows [][]string
		for _, f := range cfg.Favorites {
			rows = append(rows, []string{f.Name, f.PLZ, f.Station})
		}
		output.Table(headers, rows)
		return nil
	},
}

var favoritesAddCmd = &cobra.Command{
	Use:   "add <name> <location>",
	Short: "Add a favorite location",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		location := args[1]

		plz, err := geo.ParsePLZ(location)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		cfg, err := config.LoadDefault("meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		cfg.AddFavorite(config.Favorite{Name: name, PLZ: plz})
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Added %q (PLZ: %s) to favorites.\n", name, plz)
		return nil
	},
}

var favoritesDefaultCmd = &cobra.Command{
	Use:   "default [location]",
	Short: "Set or show the default location",
	Long: `Set or show the default location used when no location argument is given.

Without arguments, shows the current default location.
With an argument, sets the default location (place name, PLZ, or coordinates).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if len(args) == 0 {
			if cfg.DefaultLocation == "" {
				fmt.Println("No default location set. Use `meteoswiss favorites default <location>` to set one.")
			} else {
				fmt.Printf("Default location: %s\n", cfg.DefaultLocation)
			}
			return nil
		}

		cfg.DefaultLocation = args[0]
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Default location set to %q.\n", args[0])
		return nil
	},
}

var favoritesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a favorite location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		cfg.RemoveFavorite(args[0])
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Removed %q from favorites.\n", args[0])
		return nil
	},
}
