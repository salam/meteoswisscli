package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

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
	combined, _, err := listRadarAndForecastFrames(client, 24)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if !output.IsInteractive() {
		output.JSON(map[string]any{"frames": combined, "count": len(combined), "source": source.MeteoSwiss})
		return nil
	}

	output.Section(fmt.Sprintf("Radar Frames (%d)", len(combined)))
	for _, f := range combined {
		label := formatFrameLabel(f.Timestamp)
		src := f.Source
		if f.URL != "" {
			fmt.Printf("  %s  [%s]  %s\n", label, src, f.URL)
		} else {
			fmt.Printf("  %s  [%s]\n", label, src)
		}
	}
	fmt.Printf("\n%s\n", source.MeteoSwiss)
	return nil
}

func renderRadarASCII() error {
	client := api.NewClientWithCache(Lang, ResponseCache)
	combined, incaVersion, err := listRadarAndForecastFrames(client, radarFrames)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	// Limit to requested number of frames
	if radarFrames > 0 && len(combined) > radarFrames {
		combined = combined[len(combined)-radarFrames:]
	}

	if len(combined) == 0 {
		output.Error("no radar frames available")
		os.Exit(1)
	}

	for i, frame := range combined {
		label := formatFrameLabel(frame.Timestamp)
		fmt.Printf("Fetching radar frame %d/%d (%s)...\n", i+1, len(combined), label)

		var grid *output.RadarGrid

		if frame.Source == "inca" && frame.INCATs != "" {
			// Fetch INCA frame
			incaFrame, err := client.GetINCAFrame(incaVersion, frame.INCATs)
			if err != nil {
				output.Error(fmt.Sprintf("fetch INCA frame %s: %s", frame.Timestamp, err))
				continue
			}
			grid = incaFrameToGrid(incaFrame)
		} else {
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

			grid, err = output.ExtractRadarGrid(tmpFile.Name())
			os.Remove(tmpFile.Name())
			if err != nil {
				output.Error(fmt.Sprintf("parse radar data: %s", err))
				continue
			}
		}

		output.Section(fmt.Sprintf("Precipitation Radar — %s", label))
		fmt.Print(output.RenderRadarASCII(grid, radarWidth, output.NoColor, !radarNoBorder, !radarNoLakes))

		if i < len(combined)-1 {
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
	combined, incaVersion, err := listRadarAndForecastFrames(client, nFrames)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if len(combined) == 0 {
		output.Error("no radar frames available")
		os.Exit(1)
	}

	var iFrames []output.InteractiveFrame
	for i, frame := range combined {
		label := formatFrameLabel(frame.Timestamp)
		fmt.Printf("\rFetching radar frame %d/%d (%s)...", i+1, len(combined), label)

		var grid *output.RadarGrid

		if frame.Source == "inca" && frame.INCATs != "" {
			incaFrame, err := client.GetINCAFrame(incaVersion, frame.INCATs)
			if err != nil {
				output.Error(fmt.Sprintf("fetch INCA frame %s: %s", frame.Timestamp, err))
				continue
			}
			grid = incaFrameToGrid(incaFrame)
		} else {
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

			grid, err = output.ExtractRadarGrid(tmpFile.Name())
			os.Remove(tmpFile.Name())
			if err != nil {
				output.Error(fmt.Sprintf("parse radar data: %s", err))
				continue
			}
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

// incaFrameToGrid converts an INCA frame to a RadarGrid for rendering.
func incaFrameToGrid(frame *api.INCAFrame) *output.RadarGrid {
	return &output.RadarGrid{
		Rows:   frame.Rows,
		Cols:   frame.Cols,
		Data:   frame.Data,
		MinLat: 43.629,
		MaxLat: 49.363,
		MinLon: 3.169,
		MaxLon: 12.462,
	}
}

// radarAndForecastFrame combines HDF5 and INCA frame metadata.
type radarAndForecastFrame struct {
	Timestamp  string // "2006-01-02 15:04"
	URL        string // HDF5 URL (empty for INCA)
	INCATs     string // INCA timestamp "20060102_1504" (empty for HDF5)
	IsForecast bool
	Source     string // "hdf5" or "inca"
}

// listRadarAndForecastFrames merges HDF5 past + INCA present/future frames.
func listRadarAndForecastFrames(client *api.Client, hdf5Limit int) ([]radarAndForecastFrame, string, error) {
	now := time.Now().UTC()

	// Get HDF5 frames
	hdf5Frames, err := client.ListRadarFrames(hdf5Limit)
	if err != nil {
		return nil, "", fmt.Errorf("list HDF5 frames: %w", err)
	}

	// Try to get INCA version + timestamps
	incaVersion, incaErr := client.GetINCAVersion()

	var combined []radarAndForecastFrame
	seen := make(map[string]bool)

	// Add HDF5 frames
	for _, f := range hdf5Frames {
		combined = append(combined, radarAndForecastFrame{
			Timestamp: f.Timestamp,
			URL:       f.URL,
			Source:    "hdf5",
		})
		seen[f.Timestamp] = true
	}

	// Add INCA frames (only those not already covered by HDF5)
	if incaErr == nil && incaVersion != "" {
		incaTimestamps, err := client.ListINCATimestamps(incaVersion, 0)
		if err == nil {
			for _, ts := range incaTimestamps {
				// Convert INCA timestamp to display format
				t, parseErr := time.Parse("20060102_1504", ts)
				if parseErr != nil {
					continue
				}
				display := t.Format("2006-01-02 15:04")
				if seen[display] {
					continue
				}
				isForecast := t.After(now.Add(-2 * time.Minute))
				combined = append(combined, radarAndForecastFrame{
					Timestamp:  display,
					INCATs:     ts,
					IsForecast: isForecast,
					Source:     "inca",
				})
				seen[display] = true
			}
		}
	}

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Timestamp < combined[j].Timestamp
	})

	return combined, incaVersion, nil
}

// formatFrameLabel returns a display label with forecast/now annotation.
func formatFrameLabel(timestamp string) string {
	t, err := time.Parse("2006-01-02 15:04", timestamp)
	if err != nil {
		return timestamp
	}
	diff := time.Since(t)
	if diff < -30*time.Second {
		// Future
		mins := int((-diff).Minutes())
		if mins < 60 {
			return fmt.Sprintf("%s (+%dmin forecast)", timestamp, mins)
		}
		return fmt.Sprintf("%s (+%dh%02dmin forecast)", timestamp, mins/60, mins%60)
	}
	if diff < 5*time.Minute {
		return fmt.Sprintf("%s (now)", timestamp)
	}
	mins := int(diff.Minutes())
	if mins < 60 {
		return fmt.Sprintf("%s (-%dmin)", timestamp, mins)
	}
	return fmt.Sprintf("%s (-%dh%02dmin)", timestamp, mins/60, mins%60)
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
