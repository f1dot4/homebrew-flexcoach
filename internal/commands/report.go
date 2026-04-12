package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func NewReportCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "View training reports",
	}

	cmd.AddCommand(newReportListCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newReportShowCmd(rootCfg, resolvedCtx))

	return cmd
}

func newReportListCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent training reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/reports", nil)
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

			reports, ok := data["reports"].([]interface{})
			if !ok || len(reports) == 0 {
				fmt.Println("No training reports found.")
				return nil
			}

			fmt.Printf("📊 Found %d recent training reports:\n\n", len(reports))
			fmt.Printf("%-8s | %-10s | %-22s | %-10s\n", "Type", "Adherence", "Period", "Report ID")
			fmt.Println("---------+------------+------------------------+-----------")

			for _, r := range reports {
				report := r.(map[string]interface{})
				
				adherence := "N/A"
				if a, ok := report["adherence"].(float64); ok {
					adherence = fmt.Sprintf("%.0f%%", a)
				}
				
				period := fmt.Sprintf("%s to %s", report["start_date"], report["end_date"])
				
				fmt.Printf("%-8v | %-10s | %-22s | %-10v\n",
					report["type"],
					adherence,
					period,
					report["report_id"].(string)[:8])
			}

			fmt.Println("\nRun 'flexcli profile stats report show <id>' to see detailed results.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newReportShowCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "show [report-id]",
		Short: "Show detailed training report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)
			reportID := args[0]

			resp, err := client.Request("GET", fmt.Sprintf("/api/reports/%s", reportID), nil)
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

			formattedText, ok := data["formatted_text"].(string)
			if !ok {
				return fmt.Errorf("failed to get formatted report text")
			}

			// Convert Markdown-ish bold to terminal bold if needed, 
			// or just print as is since the user asked for consistency
			// We'll do a simple replacement for ** -> empty string for terminal
			output := formattedText
			output = strings.ReplaceAll(output, "**", "")
			
			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}
