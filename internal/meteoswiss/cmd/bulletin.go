package cmd

import (
	"fmt"
	"os"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/output"
	"github.com/salam/swissmeteocli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	bulletinRegion  string
	bulletinOutlook bool
)

func init() {
	rootCmd.AddCommand(bulletinCmd)
	bulletinCmd.Flags().StringVar(&bulletinRegion, "region", "north", "Region: north, south, west")
	bulletinCmd.Flags().BoolVar(&bulletinOutlook, "outlook", false, "Show extended outlook instead of today's report")
}

var bulletinCmd = &cobra.Command{
	Use:   "bulletin",
	Short: "Weather forecast bulletin (prose text)",
	Long: `Show the MeteoSwiss weather forecast in prose text.

Regions:
  --region north   Northern Switzerland & Graubünden (default)
  --region south   Southern Switzerland (Ticino, Engadin)
  --region west    Western Switzerland

Types:
  Default: today's weather report (updated ~05:00, ~12:00)
  --outlook: extended forecast outlook (updated ~22:00)`,
	Example: `  meteoswiss bulletin --region south
  meteoswiss bulletin --outlook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		region := api.BulletinRegion(bulletinRegion)
		switch region {
		case api.RegionNorth, api.RegionSouth, api.RegionWest:
		default:
			output.Error(fmt.Sprintf("unknown region %q — use north, south, or west", bulletinRegion))
			os.Exit(1)
		}

		bulletinType := api.BulletinReport
		if bulletinOutlook {
			bulletinType = api.BulletinOutlook
		}

		client := api.NewClientWithCache(Lang, ResponseCache)
		bulletin, err := client.GetBulletinText(bulletinType, region)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"type":    string(bulletin.Type),
				"region":  string(bulletin.Region),
				"lang":    bulletin.Lang,
				"version": bulletin.Version,
				"text":    bulletin.Text,
				"source":  source.MeteoSwiss,
			})
			return nil
		}

		typeName := "Weather Report"
		if bulletinOutlook {
			typeName = "Weather Outlook"
		}
		output.Section(fmt.Sprintf("%s — %s", typeName, bulletinRegion))
		fmt.Println()
		fmt.Println(bulletin.Text)
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
