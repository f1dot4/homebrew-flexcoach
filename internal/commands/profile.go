package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/f1dot4/flexcli/internal/api"

	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func NewProfileCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "User Profile Hub",
	}

	cmd.AddCommand(newProfileGetCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newProfileDeleteCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newProfileInsightsCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newBodyCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPreferencesCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewStatsCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewGoalCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewConstraintCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewProfileDataCmd(rootCfg, resolvedCtx))

	return cmd
}

func newProfileDeleteCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete user profile and all data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Print("⚠️  WARNING: This will permanently delete all your data. This action is IRREVERSIBLE.\n")
				fmt.Print("Are you sure you want to continue? (y/N): ")
				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)
			resp, err := client.Request("DELETE", "/api/profile", nil)
			if err != nil {
				return err
			}

			fmt.Printf("✅ %s\n", resp.Message)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

func newBodyCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "body",
		Short: "Body metrics and thresholds",
	}

	cmd.AddCommand(newVitalsCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewThresholdCmd(rootCfg, resolvedCtx))

	return cmd
}

func newPreferencesCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "preferences",
		Aliases: []string{"pref"},
		Short:   "Manage preferences (expert settings, custom list)",
	}

	cmd.AddCommand(newPreferencesExpertCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPreferencesCustomCmd(rootCfg, resolvedCtx))

	return cmd
}
func newProfileGetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "View full profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/profile", nil)
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

			// Debug: fmt.Printf("DEBUG: Raw API Data: %+v\n", data)

			fmt.Printf("👤 Profile for %s (ID: %s)\n", data["name"], data["user_id"])
			fmt.Printf("  • Birthdate: %v\n", data["birthdate"])
			fmt.Printf("  • Sex:       %v\n", data["sex"])
			fmt.Printf("  • Timezone:  %v\n", data["timezone"])
			fmt.Printf("  • Plan Time: %v\n", data["daily_plan_time"])
			fmt.Printf("  • Insight Time: %v\n", data["weekly_insight_time"])
                        if data["sleep_log_enabled"] == true {
                                fmt.Printf("  • Sleep Log:  ✅ (%v)\n", data["sleep_log_time"])
                        }
			if data["weight_kg"] != nil || data["height_cm"] != nil || data["bmi"] != nil {
				fmt.Println("\n⚖️ Body Vitals")
				if weight, ok := data["weight_kg"].(float64); ok {
					fmt.Printf("  • Weight:    %.1f kg\n", weight)
				} else if data["weight_kg"] != nil {
					fmt.Printf("  • Weight:    %v kg (type: %T)\n", data["weight_kg"], data["weight_kg"])
				}

				if height, ok := data["height_cm"].(float64); ok {
					fmt.Printf("  • Height:    %.0f cm\n", height)
				} else if data["height_cm"] != nil {
					fmt.Printf("  • Height:    %v cm (type: %T)\n", data["height_cm"], data["height_cm"])
				}

				if bmi, ok := data["bmi"].(float64); ok {
					fmt.Printf("  • BMI:       %.1f\n", bmi)
				} else if data["bmi"] != nil {
					fmt.Printf("  • BMI:       %v (type: %T)\n", data["bmi"], data["bmi"])
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newVitalsCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vitals",
		Short: "Manage body vitals (weight, height, sex, birthdate)",
	}

	cmd.AddCommand(newVitalsGetCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newVitalsSetCmd(rootCfg, resolvedCtx))

	return cmd
}

func newVitalsGetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "View current body vitals",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/profile", nil)
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

			fmt.Println("⚖️ Current Body Vitals")
			if weight, ok := data["weight_kg"].(float64); ok {
				fmt.Printf("  • Weight:    %.1f kg\n", weight)
			}
			if height, ok := data["height_cm"].(float64); ok {
				fmt.Printf("  • Height:    %.0f cm\n", height)
			}
			fmt.Printf("  • Sex:       %v\n", data["sex"])
			fmt.Printf("  • Birthdate: %v\n", data["birthdate"])
			if bmi, ok := data["bmi"].(float64); ok {
				fmt.Printf("  • BMI:       %.1f\n", bmi)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newVitalsSetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var weight, height float64
	var sex, birthdate string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update body vitals",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			updates := make(map[string]interface{})
			if cmd.Flags().Changed("weight") {
				updates["weight"] = weight
			}
			if cmd.Flags().Changed("height") {
				updates["height"] = height
			}
			if cmd.Flags().Changed("sex") {
				updates["sex"] = sex
			}
			if cmd.Flags().Changed("birthdate") {
				updates["birthdate"] = birthdate
			}

			if len(updates) == 0 {
				return fmt.Errorf("no updates provided. Use flags --weight, --height, etc.")
			}

			resp, err := client.Request("POST", "/api/profile/body", updates)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Printf("{\"success\": true, \"message\": \"%s\"}\n", resp.Message)
				return nil
			}

			fmt.Println("✅ Body vitals updated successfully.")
			return nil
		},
	}

	cmd.Flags().Float64Var(&weight, "weight", 0, "Weight in kg")
	cmd.Flags().Float64Var(&height, "height", 0, "Height in cm")
	cmd.Flags().StringVar(&sex, "sex", "", "Sex (male/female/other)")
	cmd.Flags().StringVar(&birthdate, "birthdate", "", "Birthdate (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")

	return cmd
}


func newProfileInsightsCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insights",
		Short: "View AI insights (coach or sleep)",
	}

	cmd.AddCommand(newProfileInsightsCoachCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newProfileInsightsSleepCmd(rootCfg, resolvedCtx))

	return cmd
}

func newProfileInsightsCoachCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var force bool
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "coach",
		Short: "View latest AI coaching insights",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			path := "/api/profile/insights"
			if force {
				path += "?force=true"
			}

			resp, err := client.Request("GET", path, nil)
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

			fmt.Println("🤖 AI Coaching Insights")

			insight, _ := data["insight"].(string)
			createdAt, _ := data["created_at"].(string)

			fmt.Printf("   (Cached from %v)\n", createdAt)
			fmt.Println()
			fmt.Printf("%v\n", insight)

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force regeneration of insights")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")

	return cmd
}

func newProfileInsightsSleepCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool
	var forceRegen bool

	cmd := &cobra.Command{
		Use:   "sleep",
		Short: "View latest sleep investigation insights",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)
			endpoint := "/api/reports/sleep-investigation"
			if forceRegen {
				endpoint += "?force=true"
			}
			resp, err := client.Request("GET", endpoint, nil)
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

			report, ok := data["report"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("failed to parse report data")
			}

			ai := report["ai_analysis"].(map[string]interface{})

			fmt.Println("🌙 SLEEP INVESTIGATION REPORT")

			createdAt, _ := report["created_at"].(string)
			fmt.Printf("   (Cached from %v)\n", createdAt)

			fmt.Println(strings.Repeat("=", 30))

			printField := func(label, key string) {
				val, ok := ai[key]
				if !ok || val == nil {
					val = "N/A"
				}
				fmt.Printf("\n%s:\n%v\n", label, val)
			}

			printField("ASSESSMENT", "sleep_quality_assessment")
			printField("REGULARITY (SRI)", "sleep_regularity_assessment")

			fmt.Println("\nTOP DRIVERS:")
			if drivers, ok := ai["top_drivers"].([]interface{}); ok {
				for _, d := range drivers {
					driver := d.(map[string]interface{})
					name := driver["name"]
					if name == nil {
						name = "Unknown"
					}

					direction := driver["direction"]
					if direction == nil {
						direction = driver["tau"]
					}
					if direction == nil {
						direction = "N/A"
					}

					confidence := driver["confidence"]
					if confidence == nil {
						confidence = "N/A"
					}

					fmt.Printf("  • %-15s: %-10s (Confidence: %s)\n",
						name, direction, confidence)
				}
			} else {
				fmt.Println("  No significant drivers identified yet.")
			}

			printField("SUGGESTED EXPERIMENT", "experiment_suggestion")
			printField("NOTE", "wearable_caveat_note")

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&forceRegen, "force", false, "Regenerate report even if one exists for today")
	return cmd
}
