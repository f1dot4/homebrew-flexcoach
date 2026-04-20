package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func NewSleepCmd(rootCfg *config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sleep",
		Short: "Manage sleep logs and investigation reports",
	}

	cmd.AddCommand(NewSleepLogCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewSleepGetCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewSleepListCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(NewSleepReportCmd(rootCfg, resolvedCtx))

	return cmd
}

func NewSleepLogCmd(rootCfg *config.Config, resolvedCtx *config.Context) *cobra.Command {
	var alcohol int
	var caffeine string
	var meal bool
	var restedness int
	var notes string
	var date string

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Submit a daily sleep log",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.Server, resolvedCtx.Key)

			payload := map[string]interface{}{
				"alcohol_units":        alcohol,
				"last_caffeine_bucket": caffeine,
				"late_heavy_meal":      meal,
				"subjective_restedness": restedness,
				"notes":                notes,
			}
			if date != "" {
				payload["date"] = date
			}

			resp, err := client.Post("/sleep-log", payload)
			if err != nil {
				return err
			}

			fmt.Println(resp.Message)
			return nil
		},
	}

	cmd.Flags().IntVar(&alcohol, "alcohol", 0, "Alcohol units consumed")
	cmd.Flags().StringVar(&caffeine, "caffeine", "before_noon", "Last caffeine bucket (before_noon, before_2pm, before_5pm, after_5pm)")
	cmd.Flags().BoolVar(&meal, "meal", false, "Had a heavy meal after 7 PM")
	cmd.Flags().IntVar(&restedness, "restedness", 3, "Subjective restedness (1-5)")
	cmd.Flags().StringVar(&notes, "notes", "", "Optional notes")
	cmd.Flags().StringVar(&date, "date", "", "Date (YYYY-MM-DD), defaults to today")

	return cmd
}

func NewSleepGetCmd(rootCfg *config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "get [date]",
		Short: "Get a sleep log for a specific date",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dateStr := "today"
			if len(args) > 0 {
				dateStr = args[0]
			}

			client := api.NewClient(resolvedCtx.Server, resolvedCtx.Key)
			resp, err := client.Get(fmt.Sprintf("/sleep-log/%s", dateStr))
			if err != nil {
				return err
			}

			if asJSON {
				data, _ := json.MarshalIndent(resp.Data, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			logData, ok := resp.Data.(map[string]interface{})["log"].(map[string]interface{})
			if !ok || logData == nil {
				fmt.Printf("No sleep log found for %s\n", dateStr)
				return nil
			}

			fmt.Printf("Sleep Log for %v:\n", logData["date"])
			fmt.Printf("  Alcohol Units: %v\n", logData["alcohol_units"])
			fmt.Printf("  Last Caffeine: %v\n", logData["last_caffeine_bucket"])
			fmt.Printf("  Late Heavy Meal: %v\n", logData["late_heavy_meal"])
			fmt.Printf("  Subjective Restedness: %v/5\n", logData["subjective_restedness"])
			if logData["notes"] != nil && logData["notes"] != "" {
				fmt.Printf("  Notes: %v\n", logData["notes"])
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func NewSleepListCmd(rootCfg *config.Config, resolvedCtx *config.Context) *cobra.Command {
	var days int
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent sleep logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.Server, resolvedCtx.Key)
			resp, err := client.Get(fmt.Sprintf("/sleep-log?days=%d", days))
			if err != nil {
				return err
			}

			if asJSON {
				data, _ := json.MarshalIndent(resp.Data, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			logs, ok := resp.Data.(map[string]interface{})["logs"].([]interface{})
			if !ok || len(logs) == 0 {
				fmt.Println("No sleep logs found.")
				return nil
			}

			fmt.Printf("Recent Sleep Logs (%d days):\n", days)
			fmt.Printf("%-12s %-8s %-12s %-6s %-10s\n", "Date", "Alcohol", "Caffeine", "Meal", "Restedness")
			fmt.Println(strings.Repeat("-", 55))

			for _, l := range logs {
				log := l.(map[string]interface{})
				fmt.Printf("%-12s %-8v %-12v %-6v %-10v\n",
					log["date"],
					log["alcohol_units"],
					log["last_caffeine_bucket"],
					log["late_heavy_meal"],
					log["subjective_restedness"],
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 7, "Number of days to list")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func NewSleepReportCmd(rootCfg *config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a sleep investigation report",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.Server, resolvedCtx.Key)
			resp, err := client.Get("/reports/sleep-investigation")
			if err != nil {
				return err
			}

			if asJSON {
				data, _ := json.MarshalIndent(resp.Data, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			report, ok := resp.Data.(map[string]interface{})["report"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("failed to parse report data")
			}

			ai := report["ai_analysis"].(map[string]interface{})
			
			fmt.Println("🌙 SLEEP INVESTIGATION REPORT")
			fmt.Println(strings.Repeat("=", 30))
			fmt.Printf("\nASSESSMENT:\n%v\n", ai["sleep_quality_assessment"])
			fmt.Printf("\nREGULARITY (SRI):\n%v\n", ai["sleep_regularity_assessment"])
			
			fmt.Println("\nTOP DRIVERS:")
			drivers := ai["top_drivers"].([]interface{})
			for _, d := range drivers {
				driver := d.(map[string]interface{})
				fmt.Printf("  • %-15s: %-10s (Confidence: %s)\n", 
					driver["name"], driver["direction"], driver["confidence"])
			}
			
			fmt.Printf("\nSUGGESTED EXPERIMENT:\n%v\n", ai["experiment_suggestion"])
			
			fmt.Printf("\nNOTE: %v\n", ai["wearable_caveat_note"])

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}
