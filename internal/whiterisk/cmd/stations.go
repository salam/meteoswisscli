package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/salam/swissmeteocli/internal/whiterisk/api"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	stationsSearch string
	stationsType   string
)

func init() {
	rootCmd.AddCommand(stationsCmd)
	stationsCmd.Flags().StringVar(&stationsSearch, "search", "", "Search by name or code")
	stationsCmd.Flags().StringVar(&stationsType, "type", "imis", "Station type: imis or study-plot")
}

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "List measurement stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)
		if stationsType == "study-plot" {
			return listStudyPlotStations(client)
		}
		return listIMISStations(client)
	},
}

func listIMISStations(client *api.Client) error {
	stations, err := client.GetIMISStations()
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	if stationsSearch != "" {
		search := strings.ToUpper(stationsSearch)
		var filtered []api.IMISStation
		for _, s := range stations {
			if strings.Contains(strings.ToUpper(s.Code), search) || strings.Contains(strings.ToUpper(s.Label), search) {
				filtered = append(filtered, s)
			}
		}
		stations = filtered
	}
	if !output.IsInteractive() {
		output.JSON(map[string]any{"stations": stations, "count": len(stations), "source": source.SLF})
		return nil
	}
	output.Section(fmt.Sprintf("IMIS Stations (%d)", len(stations)))
	headers := []string{"CODE", "NAME", "ELEVATION", "CANTON", "TYPE"}
	var rows [][]string
	for _, s := range stations {
		rows = append(rows, []string{s.Code, s.Label, fmt.Sprintf("%.0fm", s.Elevation), s.CantonCode, s.Type})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}

func listStudyPlotStations(client *api.Client) error {
	stations, err := client.GetStudyPlotStations()
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	if stationsSearch != "" {
		search := strings.ToUpper(stationsSearch)
		var filtered []api.StudyPlotStation
		for _, s := range stations {
			if strings.Contains(strings.ToUpper(s.Code), search) || strings.Contains(strings.ToUpper(s.Label), search) {
				filtered = append(filtered, s)
			}
		}
		stations = filtered
	}
	if !output.IsInteractive() {
		output.JSON(map[string]any{"stations": stations, "count": len(stations), "source": source.SLF})
		return nil
	}
	output.Section(fmt.Sprintf("Study Plot Stations (%d)", len(stations)))
	headers := []string{"CODE", "NAME", "ELEVATION", "CANTON"}
	var rows [][]string
	for _, s := range stations {
		rows = append(rows, []string{s.Code, s.Label, fmt.Sprintf("%.0fm", s.Elevation), s.CantonCode})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}
