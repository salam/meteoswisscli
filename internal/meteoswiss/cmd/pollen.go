package cmd

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/i18n"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var pollenDays int

func init() {
	rootCmd.AddCommand(pollenCmd)
	pollenCmd.Flags().IntVar(&pollenDays, "days", 5, "Number of recent days to show")
}

var pollenCmd = &cobra.Command{
	Use:   "pollen [station|location]",
	Short: "Pollen concentrations from monitoring stations",
	Long: `Show pollen concentration data from MeteoSwiss monitoring stations.

Accepts a station code (e.g. PZH), place name (e.g. Zürich), or coordinates.
Without arguments, shows the latest data from all stations.

Stations: Bern, Basel, Buchs SG, La Chaux-de-Fonds, Davos, Genève,
Jungfraujoch, Locarno, Lausanne, Lugano, Luzern, Münsterlingen,
Neuchâtel, Payerne, Sion, Zürich`,
	Example: `  meteoswiss pollen Zürich
  meteoswiss pollen PZH --days 7`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClientWithCache(Lang, ResponseCache)

		locationInput, _ := getLocationArg(args, "meteoswiss")
		if locationInput == "" {
			return showAllPollenStationsLatest(client)
		}

		// Show coordinate resolution if input looks like coordinates
		if strings.Contains(locationInput, ",") {
			if plzResolved, err := geo.ResolvePLZ(locationInput); err == nil {
				printCoordinateResolution(locationInput, plzResolved)
			}
		}

		station := resolvePollenStation(locationInput)
		if station == nil {
			output.Error(fmt.Sprintf("pollen station not found for %q. Use a station code (PZH) or city name (Zürich)", locationInput))
			os.Exit(1)
		}

		measurements, err := client.GetPollenData(station.Code)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		// Show last N days
		if pollenDays > 0 && len(measurements) > pollenDays {
			measurements = measurements[len(measurements)-pollenDays:]
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"station":      station.Code,
				"station_name": station.Name,
				"measurements": measurements,
				"source":       source.MeteoSwiss,
			})
			return nil
		}

		output.Section(fmt.Sprintf("Pollen — %s (%s)", station.Name, station.Code))
		headers := []string{i18n.T("DATE"), i18n.T("Alder"), i18n.T("Birch"), i18n.T("Hazel"), i18n.T("Beech"), i18n.T("Ash"), i18n.T("Oak"), i18n.T("Grasses")}
		var rows [][]string
		for _, m := range measurements {
			rows = append(rows, []string{
				m.Date, fmtPollen(m.Alder), fmtPollen(m.Birch), fmtPollen(m.Hazel),
				fmtPollen(m.Beech), fmtPollen(m.Ash), fmtPollen(m.Oak), fmtPollen(m.Grasses),
			})
		}
		output.Table(headers, rows)
		fmt.Println("\nUnit: No/m³ (pollen grains per cubic metre)")
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}

func fmtPollen(val string) string {
	if val == "" {
		return "-"
	}
	return val
}

func resolvePollenStation(input string) *api.PollenStation {
	input = strings.TrimSpace(input)
	upper := strings.ToUpper(input)

	// Exact station code match
	for i := range api.PollenStations {
		if strings.ToUpper(api.PollenStations[i].Code) == upper {
			return &api.PollenStations[i]
		}
	}

	// Try to resolve location via geo package, then find nearest pollen station
	resolved, err := geo.ResolvePLZ(input)
	if err != nil {
		// Fall back to name match on pollen station names
		lower := strings.ToLower(input)
		for i := range api.PollenStations {
			if strings.Contains(strings.ToLower(api.PollenStations[i].Name), lower) {
				return &api.PollenStations[i]
			}
		}
		return nil
	}

	if resolved.Location != nil {
		return findNearestPollenStation(resolved.Location.Lat, resolved.Location.Lon)
	}
	return nil
}

func findNearestPollenStation(lat, lon float64) *api.PollenStation {
	var best *api.PollenStation
	bestDist := math.MaxFloat64
	for i := range api.PollenStations {
		s := &api.PollenStations[i]
		d := pollenHaversineDist(lat, lon, s.Lat, s.Lon)
		if d < bestDist {
			bestDist = d
			best = s
		}
	}
	return best
}

func pollenHaversineDist(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func showAllPollenStationsLatest(client *api.Client) error {
	type latestEntry struct {
		Station string
		Name    string
		api.PollenMeasurement
	}

	type result struct {
		entry latestEntry
		ok    bool
	}

	results := make([]result, len(api.PollenStations))
	var wg sync.WaitGroup
	wg.Add(len(api.PollenStations))

	for i, s := range api.PollenStations {
		go func(idx int, station api.PollenStation) {
			defer wg.Done()
			measurements, err := client.GetPollenData(station.Code)
			if err != nil || len(measurements) == 0 {
				return
			}
			last := measurements[len(measurements)-1]
			results[idx] = result{
				entry: latestEntry{Station: station.Code, Name: station.Name, PollenMeasurement: last},
				ok:    true,
			}
		}(i, s)
	}
	wg.Wait()

	var entries []latestEntry
	for _, r := range results {
		if r.ok {
			entries = append(entries, r.entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Station < entries[j].Station })

	if !output.IsInteractive() {
		output.JSON(map[string]any{"stations": entries, "source": source.MeteoSwiss})
		return nil
	}

	output.Section("Pollen — All Stations (latest)")
	headers := []string{i18n.T("STATION"), i18n.T("NAME"), i18n.T("DATE"), i18n.T("Alder"), i18n.T("Birch"), i18n.T("Hazel"), i18n.T("Ash"), i18n.T("Oak"), i18n.T("Grasses")}
	var rows [][]string
	for _, e := range entries {
		rows = append(rows, []string{
			e.Station, e.Name, e.Date,
			fmtPollen(e.Alder), fmtPollen(e.Birch), fmtPollen(e.Hazel),
			fmtPollen(e.Ash), fmtPollen(e.Oak), fmtPollen(e.Grasses),
		})
	}
	output.Table(headers, rows)
	fmt.Println("\nUnit: No/m³ (pollen grains per cubic metre)")
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}
