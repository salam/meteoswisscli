package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/salam/swissmeteocli/internal/whiterisk/api"
	"github.com/salam/swissmeteocli/pkg/i18n"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	measurementType   string
	measurementPeriod int
)

func init() {
	rootCmd.AddCommand(measurementsCmd)
	measurementsCmd.Flags().StringVar(&measurementType, "type", "imis", "Station type: imis or study-plot")
	measurementsCmd.Flags().IntVar(&measurementPeriod, "period", 1, "Period in days: 1, 3, or 7")
}

var measurementsCmd = &cobra.Command{
	Use:   "measurements [station]",
	Short: "Snow and weather measurements",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		station := strings.ToUpper(args[0])
		client := api.NewClient(Lang)
		if measurementType == "study-plot" {
			return showStudyPlotMeasurements(client, station)
		}
		return showIMISMeasurements(client, station)
	},
}

func showIMISMeasurements(client *api.Client, station string) error {
	measurements, err := client.GetIMISMeasurementsByStation(station, measurementPeriod)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	if !output.IsInteractive() {
		output.JSON(map[string]any{"station": station, "measurements": measurements, "source": source.SLF})
		return nil
	}
	output.Section(fmt.Sprintf("IMIS Station: %s", station))
	headers := []string{i18n.T("TIME"), i18n.T("TEMP"), i18n.T("HUMIDITY"), i18n.T("SNOW"), i18n.T("WIND"), i18n.T("GUSTS"), i18n.T("DIR")}
	var rows [][]string
	for _, m := range measurements {
		rows = append(rows, []string{
			m.MeasureDate,
			fmtFloat(m.TA30MinMean, "°C"),
			fmtFloat(m.RH30MinMean, "%"),
			fmtFloat(m.HS, " cm"),
			fmtFloat(m.VW30MinMean, " m/s"),
			fmtFloat(m.VW30MinMax, " m/s"),
			fmtFloat(m.DW30MinMean, "°"),
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}

func showStudyPlotMeasurements(client *api.Client, station string) error {
	measurements, err := client.GetStudyPlotMeasurementsByStation(station, measurementPeriod)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	if !output.IsInteractive() {
		output.JSON(map[string]any{"station": station, "measurements": measurements, "source": source.SLF})
		return nil
	}
	output.Section(fmt.Sprintf("Study Plot Station: %s", station))
	headers := []string{i18n.T("TIME"), i18n.T("SNOW HEIGHT"), i18n.T("NEW SNOW 24h"), i18n.T("WATER EQ")}
	var rows [][]string
	for _, m := range measurements {
		rows = append(rows, []string{
			m.MeasureDate,
			fmtFloat(m.HS, " cm"),
			fmtFloat(m.HN1D, " cm"),
			fmtFloat(m.HNW1D, " mm"),
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}

func fmtFloat(v *float64, unit string) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.1f%s", *v, unit)
}
