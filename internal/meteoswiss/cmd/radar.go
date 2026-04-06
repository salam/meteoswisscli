package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	radarASCII       bool
	radarWidth       int
	radarFrames      int
	radarSave        string
	radarList        bool
	radarInteractive bool
	radarNoBorder    bool
	radarNoLakes     bool
)

func init() {
	rootCmd.AddCommand(radarCmd)
	radarCmd.Flags().BoolVar(&radarASCII, "ascii", false, "Render radar as ASCII art in terminal")
	radarCmd.Flags().IntVar(&radarWidth, "width", 120, "ASCII art width in columns")
	radarCmd.Flags().IntVar(&radarFrames, "frames", 1, "Number of time frames to show (--ascii mode)")
	radarCmd.Flags().StringVar(&radarSave, "save", "", "Save HDF5 radar file to path")
	radarCmd.Flags().BoolVar(&radarList, "list", false, "List available radar frames with timestamps")
	radarCmd.Flags().BoolVar(&radarInteractive, "interactive", false, "Interactive mode: scroll through radar frames with arrow keys")
	radarCmd.Flags().BoolVar(&radarNoBorder, "no-border", false, "Hide Swiss border outline")
	radarCmd.Flags().BoolVar(&radarNoLakes, "no-lakes", false, "Hide lake outlines")
}

var radarCmd = &cobra.Command{
	Use:   "radar [rain|cloud|satellite]",
	Short: "Weather radar and satellite images",
	Long: `View rain radar, cloud radar, or satellite imagery.

Default: opens in browser. Use --ascii to render precipitation in terminal.

Time series: use --frames N to show the last N radar snapshots (10-min intervals).
Use --interactive to scroll through frames with arrow keys.
Use --list to see available timestamps.`,
	Example: `  meteoswiss radar rain --ascii
  meteoswiss radar --interactive --frames 24`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		radarType := api.RadarRain
		if len(args) > 0 {
			switch args[0] {
			case "rain":
				radarType = api.RadarRain
			case "cloud":
				radarType = api.RadarCloud
			case "satellite":
				radarType = api.RadarSatellite
			default:
				output.Error(fmt.Sprintf("unknown radar type %q — use rain, cloud, or satellite", args[0]))
				os.Exit(1)
			}
		}

		// Interactive and list modes only work for rain (HDF5 frame data)
		if (radarInteractive || radarList) && radarType != api.RadarRain {
			output.Error("Interactive/list mode only available for rain radar (HDF5 open data)")
			fmt.Fprintf(os.Stderr, "Opening %s in browser instead...\n", radarType)
			return output.OpenBrowser(api.GetRadarBrowserURL(radarType))
		}

		if radarList {
			return listRadarFrames()
		}

		if radarInteractive {
			return renderRadarInteractive()
		}

		if radarASCII && radarType != api.RadarRain {
			return renderSatelliteASCII(radarType)
		}

		if radarASCII {
			return renderRadarASCII()
		}

		if radarSave != "" && radarType != api.RadarRain {
			return saveSatelliteImage(radarType)
		}

		if radarSave != "" {
			return saveRadarFile()
		}

		// Default: open in browser
		url := api.GetRadarBrowserURL(radarType)
		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"type":   string(radarType),
				"url":    url,
				"source": source.MeteoSwiss,
			})
			return nil
		}

		fmt.Printf("Opening %s radar in browser...\n", radarType)
		if err := output.OpenBrowser(url); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", url)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}

func listRadarFrames() error {
	client := api.NewClientWithCache(Lang, ResponseCache)
	frames, err := client.ListRadarFrames(24) // last 24 frames = ~4 hours
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if !output.IsInteractive() {
		output.JSON(map[string]any{"frames": frames, "count": len(frames), "source": source.MeteoSwiss})
		return nil
	}

	output.Section(fmt.Sprintf("Radar Frames (%d)", len(frames)))
	for _, f := range frames {
		fmt.Printf("  %s  %s\n", f.Timestamp, f.URL)
	}
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func renderRadarASCII() error {
	client := api.NewClientWithCache(Lang, ResponseCache)
	frames, err := client.ListRadarFrames(radarFrames)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if len(frames) == 0 {
		output.Error("no radar frames available")
		os.Exit(1)
	}

	for i, frame := range frames {
		fmt.Printf("Fetching radar frame %d/%d (%s)...\n", i+1, len(frames), frame.Timestamp)

		// Download HDF5 to temp file
		h5data, err := client.DownloadRadarH5(frame.URL)
		if err != nil {
			output.Error(fmt.Sprintf("download frame %s: %s", frame.Timestamp, err))
			continue
		}

		tmpFile, err := os.CreateTemp("", "radar-*.h5")
		if err != nil {
			output.Error(fmt.Sprintf("create temp file: %s", err))
			continue
		}
		tmpFile.Write(h5data)
		tmpFile.Close()

		// Extract and render
		grid, err := output.ExtractRadarGrid(tmpFile.Name())
		os.Remove(tmpFile.Name())
		if err != nil {
			output.Error(fmt.Sprintf("parse radar data: %s", err))
			continue
		}

		output.Section(fmt.Sprintf("Precipitation Radar — %s", frame.Timestamp))
		fmt.Print(output.RenderRadarASCII(grid, radarWidth, output.NoColor, !radarNoBorder, !radarNoLakes))

		if i < len(frames)-1 {
			fmt.Println()
		}
	}

	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func renderRadarInteractive() error {
	nFrames := radarFrames
	if nFrames < 2 {
		nFrames = 12 // default: 2 hours of 10-min intervals
	}

	client := api.NewClientWithCache(Lang, ResponseCache)
	frames, err := client.ListRadarFrames(nFrames)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if len(frames) == 0 {
		output.Error("no radar frames available")
		os.Exit(1)
	}

	var iFrames []output.InteractiveFrame
	for i, frame := range frames {
		fmt.Printf("\rFetching radar frame %d/%d (%s)...", i+1, len(frames), frame.Timestamp)

		h5data, err := client.DownloadRadarH5(frame.URL)
		if err != nil {
			output.Error(fmt.Sprintf("download frame %s: %s", frame.Timestamp, err))
			continue
		}

		tmpFile, err := os.CreateTemp("", "radar-*.h5")
		if err != nil {
			output.Error(fmt.Sprintf("create temp file: %s", err))
			continue
		}
		tmpFile.Write(h5data)
		tmpFile.Close()

		grid, err := output.ExtractRadarGrid(tmpFile.Name())
		os.Remove(tmpFile.Name())
		if err != nil {
			output.Error(fmt.Sprintf("parse radar data: %s", err))
			continue
		}

		iFrames = append(iFrames, output.InteractiveFrame{
			Timestamp: frame.Timestamp,
			Grid:      grid,
		})
	}

	if len(iFrames) == 0 {
		output.Error("no frames could be loaded")
		os.Exit(1)
	}

	return output.InteractiveRadar(iFrames, radarWidth, output.NoColor, !radarNoBorder, !radarNoLakes)
}

func renderSatelliteASCII(radarType api.RadarType) error {
	imageURL := api.GetSatelliteImageURL(radarType)
	if imageURL == "" {
		return fmt.Errorf("no image URL for type %s", radarType)
	}

	typeName := "Satellite"
	if radarType == api.RadarCloud {
		typeName = "Cloud"
	}

	output.Section(fmt.Sprintf("%s Imagery", typeName))
	if err := output.ASCIIMap(imageURL, radarWidth); err != nil {
		return fmt.Errorf("render %s image: %w", typeName, err)
	}
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func saveSatelliteImage(radarType api.RadarType) error {
	imageURL := api.GetSatelliteImageURL(radarType)
	if imageURL == "" {
		return fmt.Errorf("no image URL for type %s", radarType)
	}

	path := radarSave
	if filepath.Ext(path) == "" {
		path += ".png"
	}

	typeName := "satellite"
	if radarType == api.RadarCloud {
		typeName = "cloud"
	}

	fmt.Printf("Saving %s image to %s...\n", typeName, path)
	if err := output.SaveImage(imageURL, path); err != nil {
		return fmt.Errorf("save %s image: %w", typeName, err)
	}
	fmt.Printf("Saved to %s\n", path)
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func saveRadarFile() error {
	client := api.NewClientWithCache(Lang, ResponseCache)
	frames, err := client.ListRadarFrames(1)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	if len(frames) == 0 {
		output.Error("no radar frames available")
		os.Exit(1)
	}

	frame := frames[0]
	fmt.Printf("Downloading radar frame %s...\n", frame.Timestamp)

	data, err := client.DownloadRadarH5(frame.URL)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	path := radarSave
	if filepath.Ext(path) == "" {
		path += ".h5"
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		output.Error(fmt.Sprintf("write file: %s", err))
		os.Exit(1)
	}
	fmt.Printf("Saved to %s (%d bytes)\n", path, len(data))
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}
