package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func NewStatsCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "View training statistics and reports",
	}

	cmd.AddCommand(newStatsDashboardCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newStatsHealthTrendsCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewReportCmd(rootCfg, resolvedCtx))

	return cmd
}

func newStatsDashboardCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "View training dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/stats/dashboard", nil)
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

			fmt.Println("📊 Training Dashboard")
			fmt.Println("======================")

			phys, _ := data["physiological_status"].(map[string]interface{})
			if phys != nil {
				fmt.Printf("\n🧬 Physiological Status:\n")
				fmt.Printf("  • Form:    %v %v (TSB: %v)\n", phys["emoji"], phys["label"], phys["tsb"])
				fmt.Printf("  • Fitness: %v (CTL) | Fatigue: %v (ATL)\n", phys["ctl"], phys["atl"])
			}

			adherence, _ := data["adherence"].(map[string]interface{})
			if adherence != nil {
				fmt.Printf("\n📋 Plan Adherence:\n")
				fmt.Printf("  • %v%% (%v/%v sessions)\n",
					adherence["adherence_percentage"],
					adherence["completed_count"],
					adherence["planned_count"])
			}

			trends, _ := data["vital_trends"].([]interface{})
			if len(trends) > 0 {
				fmt.Printf("\n❤️ Vital Trends:\n")
				for _, t := range trends {
					trend := t.(map[string]interface{})
					fmt.Printf("  • %s: %v %s (%s)\n", trend["label"], trend["current"], trend["unit"], trend["trend"])
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newStatsHealthTrendsCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool
	var days int

	cmd := &cobra.Command{
		Use:   "healthtrends",
		Aliases: []string{"health"},
		Short: "View health trends (7d vs 30d)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", fmt.Sprintf("/api/profile/health-trends?days=%d", days), nil)
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

			metrics, ok := data["metrics"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("no health trend data available")
			}

			fmt.Printf("❤️‍🩹 Health Trends (7d / 30d)\n\n")

			trendTable := []struct {
				field string
				label string
				fmt   string
			}{
				{"weight_kg", "Weight kg", "%.1f"},
				{"body_fat_percent", "Fat %", "%.1f"},
				{"muscle_mass_kg", "Muscle kg", "%.1f"},
				{"resting_heart_rate", "RHR bpm", "%.0f"},
				{"hrv_score", "HRV ms", "%.0f"},
				{"body_battery", "Battery", "%.0f"},
				{"sleep_quality_score", "Sleep Q", "%.0f"},
				{"sleep_hours", "Sleep h", "%.1f"},
				{"ctl", "CTL", "%.1f"},
				{"atl", "ATL", "%.1f"},
				{"tsb", "TSB", "%.1f"},
				{"vo2max_running", "VO2m Run", "%.1f"},
				{"vo2max_cycling", "VO2m Bike", "%.1f"},
				{"endurance_score", "Endurance", "%.0f"},
				{"hill_score", "Hill Score", "%.0f"},
				{"fitness_age", "Fit Age", "%.0f"},
				{"race_pred_5k_seconds", "5K Pred", "%.0fs"},
				{"race_pred_10k_seconds", "10K Pred", "%.0fs"},
				{"race_pred_halfmarathon_seconds", "Half Pred", "%.0fs"},
				{"race_pred_marathon_seconds", "Mara Pred", "%.0fs"},
				{"overnight_hrv_avg", "HRV Avg", "%.0f"}, 
 
                                {"training_readiness_score", "Readiness", "%.0f"}, 
                                {"total_steps", "Steps", "%.0f"}, 
                                {"avg_stress_sleep", "Stress Slp", "%.1f"},
                                {"avg_stress_pre_sleep_2h", "PreSlp Stress", "%.1f"},
                                {"overnight_hrv_5min_high", "HRV 5m High", "%.0f"},
                                {"overnight_hrv_weekly_avg", "HRV Week Avg", "%.0f"},
                                {"hydration_ml", "Hydration ml", "%.0f"},
			}

			type row struct {
				label string
				now   string
				c7    string
				c30   string
			}
			var rows []row

			for _, t := range trendTable {
				m, ok := metrics[t.field].(map[string]interface{})
				if !ok || m["current"] == nil {
					continue
				}

				nowStr := fmt.Sprintf(t.fmt, m["current"])

				formatChange := func(val interface{}, ctype interface{}) string {
					if val == nil {
						return "—"
					}
					v := val.(float64)
					sign := ""
					if v > 0 {
						sign = "+"
					}
					suffix := ""
					if ctype == "pct" {
						suffix = "%"
					}
					if v > -0.05 && v < 0.05 {
						return "0.0" + suffix
					}
					return fmt.Sprintf("%s%.1f%s", sign, v, suffix)
				}

				c7 := formatChange(m["change_7d"], m["change_type"])
				c30 := formatChange(m["change_30d"], m["change_type"])

				rows = append(rows, row{t.label, nowStr, c7, c30})
			}

			if len(rows) == 0 {
				fmt.Println("No recent health trend data available.")
				return nil
			}

			// Calculate widths
			wMetric, wNow, w7, w30 := 6, 3, 2, 3 // Headers
			for _, r := range rows {
				if len(r.label) > wMetric { wMetric = len(r.label) }
				if len(r.now) > wNow { wNow = len(r.now) }
				if len(r.c7) > w7 { w7 = len(r.c7) }
				if len(r.c30) > w30 { w30 = len(r.c30) }
			}

			// Print table
			fmt.Printf("%-*s %*s %*s %*s\n", wMetric, "Metric", wNow, "Now", w7, "7d", w30, "30d")
			fmt.Println(strings.Repeat("─", wMetric+wNow+w7+w30+3))
			for _, r := range rows {
				fmt.Printf("%-*s %*s %*s %*s\n", wMetric, r.label, wNow, r.now, w7, r.c7, w30, r.c30)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	cmd.Flags().IntVarP(&days, "days", "d", 30, "Lookback days for trend analysis")

	return cmd
}
