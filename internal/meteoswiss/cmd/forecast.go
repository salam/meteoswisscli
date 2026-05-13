package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/i18n"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var weekFlag bool

func init() {
	rootCmd.AddCommand(forecastCmd)
	forecastCmd.Flags().BoolVar(&weekFlag, "week", false, "Show 8-day forecast")
}

var forecastCmd = &cobra.Command{
	Use:   "forecast <location> [<location> ...]",
	Short: "Weather forecast for one or more locations",
	Long:  "Show weather forecast by PLZ code (e.g. 8001), place name, or lat,lon coordinates. Pass multiple locations to compare them side by side.",
	Example: `  meteoswiss forecast Zürich
  meteoswiss forecast 8001 --week
  meteoswiss forecast Onsernone Basel`,
	Args: cobra.MaximumNArgs(5),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) >= 2 {
			return runComparisonForecast(args)
		}
		return runSingleForecast(args)
	},
}

type forecastResult struct {
	input    string
	resolved *geo.ResolvedLocation
	detail   *api.PlzDetail
}

func loadForecast(input string, client *api.Client) (*forecastResult, error) {
	resolved, err := geo.ResolvePLZ(input)
	if err != nil {
		return nil, err
	}
	detail, err := client.GetForecast(resolved.PLZ)
	if err != nil {
		return nil, err
	}
	return &forecastResult{input: input, resolved: resolved, detail: detail}, nil
}

func runSingleForecast(args []string) error {
	location, err := getLocationArg(args, "meteoswiss")
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	client := api.NewClientWithCache(Lang, ResponseCache)
	r, err := loadForecast(location, client)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	printCoordinateResolution(location, r.resolved)

	if !output.IsInteractive() {
		result := map[string]any{
			"currentWeather": r.detail.CurrentWeather,
			"forecast":       r.detail.Forecast,
			"warnings":       r.detail.Warnings,
			"source":         source.MeteoSwiss,
		}
		if r.resolved.Location != nil {
			result["location"] = r.resolved.Label()
		}
		output.JSON(result)
		return nil
	}

	if r.resolved.Location != nil {
		output.Section(r.resolved.Label())
	}

	output.Section(i18n.T("Current Weather"))
	cw := r.detail.CurrentWeather
	fmt.Printf("  %s  %.1f°C  (Icon: %d)\n", cw.TimeFormatted(), cw.Temperature, cw.Icon)

	if weekFlag {
		output.Section(i18n.T("8-Day Forecast"))
	} else {
		output.Section(i18n.T("Forecast"))
	}

	days := trimForecastDays(r.detail.Forecast)
	headers := []string{i18n.T("DATE"), i18n.T("ICON"), i18n.T("MIN"), i18n.T("MAX"), i18n.T("PRECIP")}
	var rows [][]string
	for _, d := range days {
		rows = append(rows, []string{
			d.DayDate,
			api.IconDescription(d.IconDay),
			fmt.Sprintf("%.0f°C", d.TemperatureMin),
			fmt.Sprintf("%.0f°C", d.TemperatureMax),
			fmt.Sprintf("%.1f mm", d.Precipitation),
		})
	}
	output.Table(headers, rows)

	printWarnings(r.detail.Warnings, "")

	fmt.Println()
	output.Section(i18n.T("Icons"))
	seen := make(map[int]bool)
	printIcon := func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		fmt.Printf("  Icon %d: %s\n         %s\n", id, api.IconDescription(id), api.WeatherIconURL(id))
	}
	printIcon(cw.Icon)
	for _, d := range days {
		printIcon(d.IconDay)
	}

	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func runComparisonForecast(args []string) error {
	client := api.NewClientWithCache(Lang, ResponseCache)

	results := make([]*forecastResult, 0, len(args))
	for _, input := range args {
		r, err := loadForecast(input, client)
		if err != nil {
			output.Error(fmt.Sprintf("%s: %s", input, err.Error()))
			os.Exit(1)
		}
		results = append(results, r)
	}

	if !output.IsInteractive() {
		locations := make([]map[string]any, 0, len(results))
		for _, r := range results {
			entry := map[string]any{
				"currentWeather": r.detail.CurrentWeather,
				"forecast":       r.detail.Forecast,
				"warnings":       r.detail.Warnings,
			}
			if r.resolved.Location != nil {
				entry["location"] = r.resolved.Label()
			}
			locations = append(locations, entry)
		}
		output.JSON(map[string]any{
			"locations": locations,
			"source":    source.MeteoSwiss,
		})
		return nil
	}

	for i, r := range results {
		printCoordinateResolution(r.input, r.resolved)
		label := r.input
		if r.resolved.Location != nil {
			label = r.resolved.Label()
		}
		output.Section(fmt.Sprintf("[%d] %s", i+1, label))
	}

	for i, r := range results {
		output.Section(fmt.Sprintf("%s %d", i18n.T("Current Weather"), i+1))
		cw := r.detail.CurrentWeather
		fmt.Printf("  %s  %.1f°C  (Icon: %d)\n", cw.TimeFormatted(), cw.Temperature, cw.Icon)
	}

	if weekFlag {
		output.Section(i18n.T("8-Day Forecast"))
	} else {
		output.Section(i18n.T("Forecast"))
	}

	trimmed := make([][]api.ForecastDay, len(results))
	for i, r := range results {
		trimmed[i] = trimForecastDays(r.detail.Forecast)
	}
	headers, rows := buildComparisonTable(trimmed)
	output.Table(headers, rows)

	for i, r := range results {
		if len(r.detail.Warnings) > 0 {
			printWarnings(r.detail.Warnings, fmt.Sprintf(" %d", i+1))
		}
	}

	fmt.Println()
	output.Section(i18n.T("Icons"))
	seen := make(map[int]bool)
	printIcon := func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		fmt.Printf("  Icon %d: %s\n         %s\n", id, api.IconDescription(id), api.WeatherIconURL(id))
	}
	for _, r := range results {
		printIcon(r.detail.CurrentWeather.Icon)
		for _, d := range trimForecastDays(r.detail.Forecast) {
			printIcon(d.IconDay)
		}
	}

	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func trimForecastDays(days []api.ForecastDay) []api.ForecastDay {
	if !weekFlag && len(days) > 3 {
		return days[:3]
	}
	return days
}

func buildComparisonTable(perLocation [][]api.ForecastDay) ([]string, [][]string) {
	headers := []string{i18n.T("DATE")}
	for i := range perLocation {
		n := i + 1
		headers = append(headers,
			fmt.Sprintf("%s %d", i18n.T("ICON"), n),
			fmt.Sprintf("%s %d", i18n.T("MIN"), n),
			fmt.Sprintf("%s %d", i18n.T("MAX"), n),
			fmt.Sprintf("%s %d", i18n.T("PRECIP"), n),
		)
	}

	dateSeen := make(map[string]bool)
	var dates []string
	byLocation := make([]map[string]api.ForecastDay, len(perLocation))
	for i, days := range perLocation {
		byLocation[i] = make(map[string]api.ForecastDay, len(days))
		for _, d := range days {
			byLocation[i][d.DayDate] = d
			if !dateSeen[d.DayDate] {
				dateSeen[d.DayDate] = true
				dates = append(dates, d.DayDate)
			}
		}
	}
	sort.Strings(dates)

	var rows [][]string
	for _, date := range dates {
		row := []string{date}
		for i := range perLocation {
			if d, ok := byLocation[i][date]; ok {
				row = append(row,
					api.IconDescription(d.IconDay),
					fmt.Sprintf("%.0f°C", d.TemperatureMin),
					fmt.Sprintf("%.0f°C", d.TemperatureMax),
					fmt.Sprintf("%.1f mm", d.Precipitation),
				)
			} else {
				row = append(row, "—", "—", "—", "—")
			}
		}
		rows = append(rows, row)
	}
	return headers, rows
}

func printWarnings(warnings []api.Warning, suffix string) {
	if len(warnings) == 0 {
		return
	}
	output.Section(i18n.T("Warnings") + suffix)
	for _, w := range warnings {
		text := api.WarnTypeName(w.Type)
		if len(w.Links) > 0 {
			text += " — " + w.Links[0].Text
		}
		fmt.Printf("  [Level %d] %s (%s — %s)\n", w.Level, text, w.ValidFromFormatted(), w.ValidToFormatted())
	}
}
