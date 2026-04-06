package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/i18n"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var windBrowser bool

func init() {
	rootCmd.AddCommand(windCmd)
	windCmd.Flags().BoolVar(&windBrowser, "browser", false, "Open wind animation in browser instead of showing data")
}

var windCmd = &cobra.Command{
	Use:   "wind [location]",
	Short: "Wind measurements and animations",
	Long: `Show current wind measurements from nearby stations.

Accepts a station code, place name, PLZ, or coordinates.
Without arguments, shows all stations with wind data.
Use --browser to open the wind animation in the browser.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if windBrowser {
			url := "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-wind-10m"
			if !output.IsInteractive() {
				output.JSON(map[string]string{"url": url, "source": source.MeteoSwiss})
				return nil
			}
			fmt.Println("Opening wind animation in browser...")
			output.OpenBrowser(url)
			return nil
		}

		client := api.NewClient(Lang)
		measurements, err := client.GetCurrentMeasurements("")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		// Filter to stations with wind data
		var windData []api.StationMeasurement
		if len(args) > 0 {
			resolved, resolveErr := geo.ResolveStation(args[0], 10)
			if resolveErr != nil {
				output.Error(resolveErr.Error())
				os.Exit(1)
			}
			codes := make(map[string]bool)
			for _, r := range resolved {
				codes[r.Station.Code] = true
			}
			for _, m := range measurements {
				if codes[m.Station] && m.WindSpeed != "" {
					windData = append(windData, m)
				}
			}
		} else {
			for _, m := range measurements {
				if m.WindSpeed != "" {
					windData = append(windData, m)
				}
			}
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{"measurements": windData, "source": source.MeteoSwiss})
			return nil
		}

		title := i18n.T("WIND")
		if len(args) > 0 {
			title += " — " + args[0]
		}
		output.Section(title)
		headers := []string{i18n.T("STATION"), i18n.T("NAME"), i18n.T("WIND"), i18n.T("GUSTS"), i18n.T("DIR")}
		var rows [][]string
		for _, m := range windData {
			name := ""
			if s := geo.LookupStation(m.Station); s != nil {
				name = s.Name
			}
			rows = append(rows, []string{
				m.Station,
				name,
				fmtVal(m.WindSpeed, " km/h"),
				fmtVal(m.GustPeak, " km/h"),
				fmtVal(m.WindDir, "°"),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
