package cmd

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/salam/swissmeteocli/internal/whiterisk/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	stationsSearch string
	stationsType   string
	stationsNear   string
	stationsLimit  int
)

func init() {
	rootCmd.AddCommand(stationsCmd)
	stationsCmd.Flags().StringVar(&stationsSearch, "search", "", "Search by name or code")
	stationsCmd.Flags().StringVar(&stationsType, "type", "imis", "Station type: imis or study-plot")
	stationsCmd.Flags().StringVar(&stationsNear, "near", "", "Find stations near a place name, PLZ, or lat,lon")
	stationsCmd.Flags().IntVar(&stationsLimit, "limit", 10, "Number of nearby stations to show (with --near)")
}

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "List measurement stations",
	Long:  "List IMIS or study plot stations. Use --near to find stations close to a location.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)
		if stationsType == "study-plot" {
			return listStudyPlotStations(client)
		}
		return listIMISStations(client)
	},
}

type stationWithDist struct {
	station  api.IMISStation
	distance float64
}

type studyPlotWithDist struct {
	station  api.StudyPlotStation
	distance float64
}

func resolveNearCoords() (lat, lon float64, label string, err error) {
	resolved, err := geo.ResolvePLZ(stationsNear)
	if err != nil {
		return 0, 0, "", err
	}
	if resolved.Location != nil {
		return resolved.Location.Lat, resolved.Location.Lon, resolved.Label(), nil
	}
	// Fallback: shouldn't happen since ResolvePLZ always resolves to a Location now
	return 0, 0, "", fmt.Errorf("could not resolve location %q", stationsNear)
}

func haversineDist(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func listIMISStations(client *api.Client) error {
	stations, err := client.GetIMISStations()
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	// --near: sort by distance
	if stationsNear != "" {
		lat, lon, label, err := resolveNearCoords()
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		var withDist []stationWithDist
		for _, s := range stations {
			d := haversineDist(lat, lon, s.Lat, s.Lon)
			withDist = append(withDist, stationWithDist{station: s, distance: d})
		}
		sort.Slice(withDist, func(i, j int) bool { return withDist[i].distance < withDist[j].distance })

		if stationsLimit > 0 && len(withDist) > stationsLimit {
			withDist = withDist[:stationsLimit]
		}

		if !output.IsInteractive() {
			type stationResult struct {
				api.IMISStation
				DistanceKm float64 `json:"distance_km"`
			}
			results := make([]stationResult, len(withDist))
			for i, sd := range withDist {
				results[i] = stationResult{IMISStation: sd.station, DistanceKm: math.Round(sd.distance*10) / 10}
			}
			output.JSON(map[string]any{"near": label, "stations": results, "count": len(results), "source": source.SLF})
			return nil
		}

		output.Section(fmt.Sprintf("IMIS Stations near %s (%d)", label, len(withDist)))
		headers := []string{"CODE", "NAME", "ELEVATION", "CANTON", "TYPE", "DISTANCE"}
		var rows [][]string
		for _, sd := range withDist {
			s := sd.station
			rows = append(rows, []string{
				s.Code, s.Label, fmt.Sprintf("%.0fm", s.Elevation), s.CantonCode, s.Type,
				fmt.Sprintf("%.1f km", sd.distance),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	}

	// --search: filter by name/code
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

	// --near: sort by distance
	if stationsNear != "" {
		lat, lon, label, err := resolveNearCoords()
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		var withDist []studyPlotWithDist
		for _, s := range stations {
			d := haversineDist(lat, lon, s.Lat, s.Lon)
			withDist = append(withDist, studyPlotWithDist{station: s, distance: d})
		}
		sort.Slice(withDist, func(i, j int) bool { return withDist[i].distance < withDist[j].distance })

		if stationsLimit > 0 && len(withDist) > stationsLimit {
			withDist = withDist[:stationsLimit]
		}

		if !output.IsInteractive() {
			type stationResult struct {
				api.StudyPlotStation
				DistanceKm float64 `json:"distance_km"`
			}
			results := make([]stationResult, len(withDist))
			for i, sd := range withDist {
				results[i] = stationResult{StudyPlotStation: sd.station, DistanceKm: math.Round(sd.distance*10) / 10}
			}
			output.JSON(map[string]any{"near": label, "stations": results, "count": len(results), "source": source.SLF})
			return nil
		}

		output.Section(fmt.Sprintf("Study Plot Stations near %s (%d)", label, len(withDist)))
		headers := []string{"CODE", "NAME", "ELEVATION", "CANTON", "DISTANCE"}
		var rows [][]string
		for _, sd := range withDist {
			s := sd.station
			rows = append(rows, []string{
				s.Code, s.Label, fmt.Sprintf("%.0fm", s.Elevation), s.CantonCode,
				fmt.Sprintf("%.1f km", sd.distance),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	}

	// --search: filter by name/code
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
