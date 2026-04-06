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

var bulletinPDF bool

func init() {
	rootCmd.AddCommand(bulletinCmd)
	bulletinCmd.Flags().BoolVar(&bulletinPDF, "pdf", false, "Download bulletin as PDF")
}

var bulletinCmd = &cobra.Command{
	Use:   "bulletin [location]",
	Short: "Avalanche bulletin",
	Long:  "Show avalanche danger ratings. Filter by region name, region ID, or lat,lon.",
	Example: `  whiterisk bulletin Davos
  whiterisk bulletin --pdf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClientWithCache(Lang, ResponseCache)

		if bulletinPDF {
			url := client.GetBulletinPDFURL()
			if !output.IsInteractive() {
				output.JSON(map[string]string{"pdf_url": url, "source": source.SLF})
				return nil
			}
			fmt.Println("Opening bulletin PDF in browser...")
			if err := output.OpenBrowser(url); err != nil {
				fmt.Printf("Download: %s\n", url)
			}
			return nil
		}

		result, err := client.GetBulletin()
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		bulletins := result.Bulletins
		if len(args) > 0 {
			search := args[0]
			bulletins = filterBulletins(bulletins, search)
			if len(bulletins) == 0 {
				output.Error(fmt.Sprintf("no bulletin found for %q", search))
				os.Exit(1)
			}
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{"bulletins": bulletins, "source": source.SLF})
			return nil
		}

		for _, b := range bulletins {
			regions := make([]string, len(b.Regions))
			for i, r := range b.Regions {
				regions[i] = fmt.Sprintf("%s (%s)", r.Name, r.RegionID)
			}

			output.Section(i18n.T("Avalanche Bulletin"))
			fmt.Printf("  Regions: %s\n", strings.Join(regions, ", "))
			fmt.Printf("  Valid:   %s → %s\n", b.ValidTime.StartTime, b.ValidTime.EndTime)

			if len(b.DangerRatings) > 0 {
				for _, dr := range b.DangerRatings {
					elev := ""
					if ub := dr.Elevation.UpperBoundStr(); ub != "" {
						elev = fmt.Sprintf(" (below %s)", ub)
					}
					if lb := dr.Elevation.LowerBoundStr(); lb != "" {
						elev = fmt.Sprintf(" (above %s)", lb)
					}
					fmt.Printf("  Danger:  %s%s\n", api.DangerLevelDisplay(dr.MainValue), elev)
				}
			}

			if len(b.AvalancheProblems) > 0 {
				fmt.Println("  Problems:")
				for _, p := range b.AvalancheProblems {
					aspects := strings.Join(p.Aspects, "/")
					elev := ""
					if lb := p.Elevation.LowerBoundStr(); lb != "" {
						elev = fmt.Sprintf(" above %s", lb)
					}
					fmt.Printf("    - %s — %s%s\n", p.ProblemType, aspects, elev)
				}
			}

			if b.SnowpackStructure != nil && b.SnowpackStructure.Comment != "" {
				fmt.Printf("  Snowpack: %s\n", b.SnowpackStructure.Comment)
			}
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	},
}

func filterBulletins(bulletins []api.Bulletin, search string) []api.Bulletin {
	search = strings.ToLower(search)
	var matched []api.Bulletin
	for _, b := range bulletins {
		for _, r := range b.Regions {
			if strings.ToLower(r.RegionID) == strings.ToLower(search) ||
				strings.Contains(strings.ToLower(r.Name), search) {
				matched = append(matched, b)
				break
			}
		}
	}
	return matched
}
