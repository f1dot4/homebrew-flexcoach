package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func newPreferencesExpertCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expert",
		Short: "Manage expert settings (SLEEP_LOG_ENABLED, sync intervals, LANGUAGE, etc.)",
	}

	cmd.AddCommand(newPreferencesExpertGetCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newPreferencesExpertSetCmd(rootCfg, resolvedCtx))

	return cmd
}

func newPreferencesExpertGetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "View current preferences",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/profile/preferences/effective", nil)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			var settings []map[string]interface{}
			if err := json.Unmarshal(resp.Data, &settings); err != nil {
				return fmt.Errorf("failed to parse settings data: %w (raw: %s)", err, string(resp.Data))
			}
			fmt.Println("⚙️ Current Preferences")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "KEY\tVALUE\tSOURCE\tDESCRIPTION")
			for _, s := range settings {
				fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", s["key"], s["value"], s["source"], s["description"])
			}
			w.Flush()
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newPreferencesExpertSetCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "set [KEY=VALUE...]",
		Short: "Update user preferences using KEY=VALUE pairs",
		Long: `Update user preferences. 
Expert settings and basic settings can also be set via KEY=VALUE positional arguments. Use KEY= to reset a setting to its system default.
Example: flexcli profile preferences expert set WITHINGS_SYNC_INTERVAL_HOURS=2 LANGUAGE=Deutsch timezone=Europe/Vienna
Example (reset): flexcli profile preferences expert set WITHINGS_SYNC_INTERVAL_HOURS=`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			updates := make(map[string]interface{})

			// Parse positional arguments KEY=VALUE
			for _, arg := range args {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid argument format '%s', expected KEY=VALUE", arg)
				}
				updates[parts[0]] = parts[1]
			}

			if len(updates) == 0 {
				return fmt.Errorf("no updates provided. Use KEY=VALUE pairs")
			}

			resp, err := client.Request("POST", "/api/profile/preferences", updates)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Printf("{\"success\": true, \"message\": \"%s\"}\n", resp.Message)
				return nil
			}

			fmt.Println("✅ Preferences updated successfully.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")

	return cmd
}
