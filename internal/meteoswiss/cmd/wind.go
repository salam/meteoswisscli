package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/i18n"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	windBrowser  bool
	windASCII    bool
	windWidth    int
	windNoBorder bool
	windNoLakes  bool
)

func init() {
	rootCmd.AddCommand(windCmd)
	windCmd.Flags().BoolVar(&windBrowser, "browser", false, "Open wind animation in browser instead of showing data")
	windCmd.Flags().BoolVar(&windASCII, "ascii", false, "Render wind map as ASCII art in terminal")
	windCmd.Flags().IntVar(&windWidth, "width", 120, "ASCII art width in columns")
	windCmd.Flags().BoolVar(&windNoBorder, "no-border", false, "Hide Swiss border outline (--ascii mode)")
	windCmd.Flags().BoolVar(&windNoLakes, "no-lakes", false, "Hide lake outlines (--ascii mode)")
}

var windCmd = &cobra.Command{
	Use:   "wind [location]",
	Short: "Wind measurements and animations",
	Long: `Show current wind measurements from nearby stations.

Accepts a station code, place name, PLZ, or coordinates.
Without arguments, shows all stations with wind data.
Use --browser to open the wind animation in the browser.`,
	Example: `  meteoswiss wind "Arosa GR"
  meteoswiss wind --browser`,
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

		if windASCII {
			client := api.NewClientWithCache(Lang, ResponseCache)
			measurements, err := client.GetCurrentMeasurements("")
			if err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			var windData []api.StationMeasurement
			for _, m := range measurements {
				if m.WindSpeed != "" {
					windData = append(windData, m)
				}
			}

			// Determine highlight location if a location argument was given
			var highlightLat, highlightLon float64
			locationInput, _ := getLocationArg(args, "meteoswiss")
			if locationInput != "" {
				resolved, resolveErr := geo.ResolveStation(locationInput, 1)
				if resolveErr == nil && len(resolved) > 0 {
					highlightLat = resolved[0].Station.Lat
					highlightLon = resolved[0].Station.Lon
				}
			}

			output.Section("Wind Map")
			fmt.Print(renderWindASCII(windData, windWidth, output.NoColor, !windNoBorder, !windNoLakes, highlightLat, highlightLon))

			// If location was specified, also show the data table below
			if locationInput != "" {
				var filteredData []api.StationMeasurement
				resolved, resolveErr := geo.ResolveStation(locationInput, 10)
				if resolveErr == nil {
					codes := make(map[string]bool)
					for _, r := range resolved {
						codes[r.Station.Code] = true
					}
					for _, m := range measurements {
						if codes[m.Station] && m.WindSpeed != "" {
							filteredData = append(filteredData, m)
						}
					}
				}

				if len(filteredData) > 0 {
					title := i18n.T("WIND") + " — " + locationInput
					output.Section(title)
					headers := []string{i18n.T("STATION"), i18n.T("NAME"), i18n.T("WIND"), i18n.T("GUSTS"), i18n.T("DIR")}
					var rows [][]string
					for _, m := range filteredData {
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
				}
			}

			fmt.Printf("\n%s\n", source.MeteoSwiss)
			return nil
		}

		client := api.NewClientWithCache(Lang, ResponseCache)
		measurements, err := client.GetCurrentMeasurements("")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		// Filter to stations with wind data
		var windData []api.StationMeasurement
		locationInput, _ := getLocationArg(args, "meteoswiss")
		if locationInput != "" {
			// Show coordinate resolution if input looks like coordinates
			if strings.Contains(locationInput, ",") {
				if plzResolved, err := geo.ResolvePLZ(locationInput); err == nil {
					printCoordinateResolution(locationInput, plzResolved)
				}
			}

			resolved, resolveErr := geo.ResolveStation(locationInput, 10)
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
		if locationInput != "" {
			title += " — " + locationInput
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
