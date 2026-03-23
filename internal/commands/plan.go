package commands

import (
	"encoding/json"
	"fmt"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func NewPlanCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Manage training plans",
	}

	cmd.AddCommand(newPlanGetCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPlanGenerateCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPlanModifyCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPlanSkipCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPlanListCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPlanActivateCmd(rootCfg, resolvedCtx))

	return cmd
}

func newPlanGenerateCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var instructions string
	var asJSON bool
	var meso bool
	var macro bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate today's plan (or meso/macro plan)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			payload := map[string]interface{}{}
			if instructions != "" {
				payload["instructions"] = instructions
			}

			planType := "daily"
			if meso {
				planType = "meso"
			} else if macro {
				planType = "macro"
			}
			payload["type"] = planType

			resp, err := client.Request("POST", "/api/plan/generate", payload)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Println(resp.Message)
			return nil
		},
	}

	cmd.Flags().StringVarP(&instructions, "instructions", "i", "", "Optional instructions for generation")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&meso, "meso", false, "Generate a meso (weekly) plan")
	cmd.Flags().BoolVar(&macro, "macro", false, "Generate a macro (4-week) plan")
	return cmd
}

func newPlanListCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool
	var statusFilter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all training plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/plans", nil)
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

			plans, ok := data["plans"].([]interface{})
			if !ok || len(plans) == 0 {
				fmt.Println("No training plans found.")
				return nil
			}

			// Filter by status if requested
			filteredPlans := []interface{}{}
			if statusFilter != "" {
				for _, p := range plans {
					plan := p.(map[string]interface{})
					if s, ok := plan["status"].(string); ok && s == statusFilter {
						filteredPlans = append(filteredPlans, p)
					}
				}
			} else {
				filteredPlans = plans
			}

			if len(filteredPlans) == 0 {
				fmt.Printf("No training plans found with status '%s'.\n", statusFilter)
				return nil
			}

			fmt.Printf("📋 Found %d training plans:\n\n", len(filteredPlans))
			fmt.Printf("%-36s | %-10s | %-10s | %s\n", "Plan ID", "Type", "Date", "Status")
			fmt.Println("-------------------------------------+------------+------------+---------")

			for _, p := range filteredPlans {
				plan := p.(map[string]interface{})
				planDate := "N/A"
				if pd, ok := plan["plan_date"].(string); ok && pd != "" {
					planDate = pd[:10] // Just the date part
				} else if sd, ok := plan["start_date"].(string); ok && sd != "" {
					planDate = sd[:10]
				}
				status := fmt.Sprintf("%v", plan["status"])
				if reason, ok := plan["skip_reason"].(string); ok && reason != "" {
					status = fmt.Sprintf("%s [%s]", status, reason)
				}
				fmt.Printf("%-36v | %-10v | %-10s | %s\n",
					plan["plan_id"],
					plan["plan_type"],
					planDate,
					status)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Filter by status (active, scheduled, inactive)")
	return cmd
}

func newPlanModifyCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var instructions string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Modify today's plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			if instructions == "" {
				return fmt.Errorf("instructions are required for modification")
			}

			payload := map[string]string{
				"instructions": instructions,
			}

			resp, err := client.Request("POST", "/api/plan/modify", payload)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Println(resp.Message)
			return nil
		},
	}

	cmd.Flags().StringVarP(&instructions, "instructions", "i", "", "Required instructions for modification")
	cmd.MarkFlagRequired("instructions")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newPlanGetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get today's plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/plan", nil)
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

			if data["plan"] == nil {
				fmt.Println("No plan found for today.")
				return nil
			}

			plan := data["plan"].(map[string]interface{})
			fmt.Printf("📅 Plan for %v\n", plan["plan_date"])

			status := fmt.Sprintf("%v", plan["status"])
			if plan["user_modifications"] != nil {
				// Simplified check for skip reason
				mods := plan["user_modifications"].(map[string]interface{})
				if history, ok := mods["history"].([]interface{}); ok && len(history) > 0 {
					last_mod := history[len(history)-1].(map[string]interface{})
					if action, ok := last_mod["action"].(string); ok && action == "skip" {
						status = fmt.Sprintf("Skipped (%v)", last_mod["reason"])
					}
				}
			}
			fmt.Printf("Status: %s\n", status)

			if genTime, ok := plan["plan_create_datetime"].(string); ok {
				fmt.Printf("Generated: %s\n", genTime)
			}

			fmt.Println("\nActivities:")
			activities, _ := plan["activities"].([]interface{})
			for _, a := range activities {
				act := a.(map[string]interface{})
				fmt.Printf("  • %v (%vm): %v\n", act["sport_type"], act["duration_minutes"], act["user_description"])

				details := ""
				if dist, ok := act["distance_km"].(float64); ok && dist > 0 {
					details += fmt.Sprintf("%.1f km", dist)
				}
				if elev, ok := act["elevation_gain_meters"].(float64); ok && elev > 0 {
					if details != "" {
						details += " | "
					}
					details += fmt.Sprintf("%.0f m ascent", elev)
				}
				if details != "" {
					fmt.Printf("    • %s\n", details)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newPlanSkipCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
        var reason string
        var asJSON bool

        cmd := &cobra.Command{
                Use:   "skip [plan-id]",
                Short: "Skip today's plan (or a specific plan by ID)",
                Args:  cobra.MaximumNArgs(1),
                RunE: func(cmd *cobra.Command, args []string) error {
                        client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

                        path := fmt.Sprintf("/api/plan/skip?reason=%s", reason)
                        if len(args) > 0 {
                                path = fmt.Sprintf("%s&plan_id=%s", path, args[0])
                        }

                        resp, err := client.Request("POST", path, nil)
                        if err != nil {
                                return err
                        }

                        if asJSON {
                                fmt.Printf("{\"success\": true, \"message\": \"%s\"}\n", resp.Message)
                                return nil
                        }

                        fmt.Println(resp.Message)
                        return nil
                },
        }

        cmd.Flags().StringVar(&reason, "reason", "other", "Reason for skipping")
        cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
        return cmd
}
func newPlanActivateCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "activate [plan-id]",
		Short: "Manually activate a training plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)
			planID := args[0]

			path := fmt.Sprintf("/api/plans/%s/activate", planID)
			resp, err := client.Request("POST", path, nil)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Println(resp.Message)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}
