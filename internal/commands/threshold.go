package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func NewThresholdCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threshold",
		Short: "Manage training thresholds",
	}

	cmd.AddCommand(newThresholdGetCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newThresholdSetCmd(rootCfg, resolvedCtx))

	return cmd
}

func newThresholdGetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get current thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/thresholds", nil)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			var data map[string]interface{}
			if err := json.Unmarshal(resp.Data, &data); err != nil {
				return err
			}

			if data["thresholds"] == nil {
				fmt.Println("No thresholds found.")
				return nil
			}

			thresholds := data["thresholds"].(map[string]interface{})
			lastSeenDates := make(map[string]time.Time)
			if lsd, ok := thresholds["last_seen_dates"].(map[string]interface{}); ok {
				for k, v := range lsd {
					if dateStr, ok := v.(string); ok {
						t, _ := time.Parse("2006-01-02", dateStr)
						lastSeenDates[k] = t
					}
				}
			}

			fmt.Println("📊 Current Training Thresholds")
			fmt.Println("------------------------------")

			today := time.Now()
			
			formatLine := func(metricKey, effectiveKey, metricName, unit string) string {
				val := thresholds[metricKey]
				eff := thresholds[effectiveKey]
				isLearned, _ := thresholds["is_"+metricName+"_learned"].(bool)
				isDerived, _ := thresholds["is_"+metricName+"_derived"].(bool)
				
				valStr := fmt.Sprintf("%v", val)
				hasManual := val != nil && valStr != "<nil>" && valStr != ""
				
				effStr := fmt.Sprintf("%v", eff)
				hasEffective := eff != nil && effStr != "<nil>" && effStr != ""
				
				if !hasEffective && !hasManual {
					return "N/A"
				}

				displayUnit := ""
				if unit != "" {
					displayUnit = " " + unit
				}

				icon := ""
				if isLearned {
					icon = " 📈"
				} else if isDerived {
					icon = " 🔢"
				}

				staleIcon := ""
				if lastSeen, ok := lastSeenDates[metricName]; ok {
					if today.Sub(lastSeen).Hours() > 24*30 {
						staleIcon = " ⚠️"
					}
				}

				if hasManual {
					// Check if history suggests a different value (staleness check)
					if staleIcon != "" && hasEffective && effStr != valStr {
						return fmt.Sprintf("%v%s%s (%v%s%s)", val, displayUnit, staleIcon, eff, displayUnit, icon)
					}
					return fmt.Sprintf("%v%s%s", val, displayUnit, staleIcon)
				}

				// No manual, show effective
				return fmt.Sprintf("%v%s%s%s", eff, displayUnit, icon, staleIcon)
			}

			fmt.Println("Running:")
			fmt.Printf("  • FTP:   %s\n", formatLine("running_ftp", "effective_running_ftp", "running_ftp", "W"))
			fmt.Printf("  • LTHR:  %s\n", formatLine("running_lthr", "effective_running_lthr", "running_lthr", "bpm"))
			fmt.Printf("  • Pace:  %s\n", formatLine("running_threshold_pace", "effective_running_threshold_pace", "running_threshold_pace", ""))

			fmt.Println("\nCycling:")
			fmt.Printf("  • FTP:   %s\n", formatLine("cycling_ftp", "effective_cycling_ftp", "cycling_ftp", "W"))
			fmt.Printf("  • LTHR:  %s\n", formatLine("cycling_lthr", "effective_cycling_lthr", "cycling_lthr", "bpm"))
			fmt.Printf("  • Pace:  %s\n", formatLine("cycling_threshold_pace", "effective_cycling_threshold_pace", "cycling_threshold_pace", ""))

			fmt.Println("\nLegend:")
			fmt.Println("  📈 = learned from history")
			fmt.Println("  🔢 = calculated via formula")
			fmt.Println("  ⚠️  = stale (not seen in 30 days)")

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newThresholdSetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var rFTP, rLTHR, cFTP, cLTHR int
	var rPace, cPace string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			update := make(map[string]interface{})
			if cmd.Flags().Changed("running-ftp") {
				update["running_ftp"] = rFTP
			}
			if cmd.Flags().Changed("running-lthr") {
				update["running_lthr"] = rLTHR
			}
			if cmd.Flags().Changed("running-pace") {
				update["running_threshold_pace"] = rPace
			}
			if cmd.Flags().Changed("cycling-ftp") {
				update["cycling_ftp"] = cFTP
			}
			if cmd.Flags().Changed("cycling-lthr") {
				update["cycling_lthr"] = cLTHR
			}
			if cmd.Flags().Changed("cycling-pace") {
				update["cycling_threshold_pace"] = cPace
			}

			if len(update) == 0 {
				return fmt.Errorf("no threshold values provided to set")
			}

			resp, err := client.Request("POST", "/api/thresholds", update)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Println("✅ Thresholds updated successfully.")
			return nil
		},
	}

	cmd.Flags().IntVar(&rFTP, "running-ftp", 0, "Running FTP (W)")
	cmd.Flags().IntVar(&rLTHR, "running-lthr", 0, "Running LTHR (bpm)")
	cmd.Flags().StringVar(&rPace, "running-pace", "", "Running Pace (e.g. 4:30/km)")
	cmd.Flags().IntVar(&cFTP, "cycling-ftp", 0, "Cycling FTP (W)")
	cmd.Flags().IntVar(&cLTHR, "cycling-lthr", 0, "Cycling LTHR (bpm)")
	cmd.Flags().StringVar(&cPace, "cycling-pace", "", "Cycling Pace (e.g. 1:20/km)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")

	return cmd
}
