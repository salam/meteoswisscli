package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/pkg/config"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(favoritesCmd)
	favoritesCmd.AddCommand(favoritesAddCmd)
	favoritesCmd.AddCommand(favoritesRemoveCmd)
	favoritesCmd.AddCommand(favoritesListCmd)
}

var favoritesCmd = &cobra.Command{
	Use:   "favorites",
	Short: "Manage saved locations and stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return favoritesListCmd.RunE(cmd, args)
	},
}

var favoritesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("whiterisk")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}
		if !output.IsInteractive() {
			output.JSON(cfg.Favorites)
			return nil
		}
		if len(cfg.Favorites) == 0 {
			fmt.Println("No favorites saved. Use `whiterisk favorites add <name> <station|region>` to add one.")
			return nil
		}
		headers := []string{"NAME", "STATION", "REGION"}
		var rows [][]string
		for _, f := range cfg.Favorites {
			rows = append(rows, []string{f.Name, f.Station, f.Region})
		}
		output.Table(headers, rows)
		return nil
	},
}

var favoritesAddCmd = &cobra.Command{
	Use:   "add <name> <station-or-region>",
	Short: "Add a favorite station or region",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		value := args[1]
		cfg, err := config.LoadDefault("whiterisk")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}
		fav := config.Favorite{Name: name}
		if len(value) <= 6 {
			fav.Station = value
		} else {
			fav.Region = value
		}
		cfg.AddFavorite(fav)
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Added %q to favorites.\n", name)
		return nil
	},
}

var favoritesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a favorite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("whiterisk")
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
