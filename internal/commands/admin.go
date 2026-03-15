package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

func NewAdminCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "System administration commands",
	}

	adminCmd.AddCommand(newAdminStatusCmd(resolvedCtx))
	adminCmd.AddCommand(newAdminUsersCmd(resolvedCtx))
	adminCmd.AddCommand(newAdminBackupCmd(resolvedCtx))
	adminCmd.AddCommand(newAdminSyncAllCmd(resolvedCtx))
	adminCmd.AddCommand(newAdminSettingsCmd(resolvedCtx))

	return adminCmd
}

func newAdminStatusCmd(ctx *config.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get system-wide status and health",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			resp, err := client.Request("GET", "/api/admin/status", nil)
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var status map[string]interface{}
			if err := json.Unmarshal(resp.Data, &status); err != nil {
				// If unmarshal fails, just print raw data
				fmt.Println(string(resp.Data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PROPERTY\tVALUE")
			fmt.Fprintf(w, "Overall Status\t%v\n", status["overall_status"])
			fmt.Fprintf(w, "Scheduler Running\t%v\n", status["scheduler_running"])
			fmt.Fprintf(w, "Total Tasks\t%v\n", status["total_tasks"])
			fmt.Fprintf(w, "Enabled Tasks\t%v\n", status["enabled_tasks"])
			fmt.Fprintf(w, "Disabled Tasks\t%v\n", status["disabled_tasks"])
			fmt.Fprintf(w, "Avg Success Rate (7d)\t%v\n", status["average_success_rate_7d"])
			fmt.Fprintf(w, "Last Updated\t%v\n", status["last_updated"])

			if failures, ok := status["recent_failures"].([]interface{}); ok && len(failures) > 0 {
				fmt.Fprintln(w, "\nRECENT FAILURES\t")
				fmt.Fprintln(w, "TASK ID\tERROR\tFAILED AT")
				for _, f := range failures {
					failure := f.(map[string]interface{})
					fmt.Fprintf(w, "%v\t%v\t%v\n", failure["task_id"], failure["error"], failure["failed_at"])
				}
			}

			w.Flush()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

func newAdminUsersCmd(ctx *config.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "users",
		Short: "List all user profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			resp, err := client.Request("GET", "/api/admin/users", nil)
			if err != nil {
				return err
			}

			var users []map[string]interface{}
			if err := json.Unmarshal(resp.Data, &users); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "USER ID\tNAME\tTIMEZONE\tCREATED AT")
			for _, u := range users {
				fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", u["user_id"], u["name"], u["timezone"], u["created_at"])
			}
			w.Flush()
			return nil
		},
	}
}

func newAdminSyncAllCmd(ctx *config.Context) *cobra.Command {
	var source string
	cmd := &cobra.Command{
		Use:   "sync-all",
		Short: "Trigger background sync for all users",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			path := "/api/admin/sync-all"
			if source != "" {
				path = fmt.Sprintf("%s?source=%s", path, source)
			}

			resp, err := client.Request("POST", path, nil)
			if err != nil {
				return err
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "Specific sync source (garmin, withings)")
	return cmd
}

func newAdminBackupCmd(ctx *config.Context) *cobra.Command {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage system backups",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List backup history",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			resp, err := client.Request("GET", "/api/admin/backups", nil)
			if err != nil {
				return err
			}

			var history []map[string]interface{}
			if err := json.Unmarshal(resp.Data, &history); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TIMESTAMP\tFILENAME\tSIZE (KB)\tSUCCESS")
			for _, h := range history {
				sizeKB := int(h["size_bytes"].(float64)) / 1024
				fmt.Fprintf(w, "%v\t%v\t%d\t%v\n", h["timestamp"], h["filename"], sizeKB, h["success"])
			}
			w.Flush()
			return nil
		},
	}

	triggerCmd := &cobra.Command{
		Use:   "create",
		Short: "Trigger immediate backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			resp, err := client.Request("POST", "/api/admin/backups/trigger", nil)
			if err != nil {
				return err
			}

			fmt.Println(resp.Message)
			return nil
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "List backup configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			resp, err := client.Request("GET", "/api/admin/backups/settings", nil)
			if err != nil {
				return err
			}

			var settings []map[string]interface{}
			if err := json.Unmarshal(resp.Data, &settings); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "KEY\tVALUE\tDESCRIPTION")
			for _, s := range settings {
				fmt.Fprintf(w, "%v\t%v\t%v\n", s["key"], s["value"], s["description"])
			}
			w.Flush()
			return nil
		},
	}

	setConfigCmd := &cobra.Command{
		Use:   "set-config [key] [value]",
		Short: "Update a backup configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			key := args[0]
			val := args[1]

			body := map[string]interface{}{
				"value": val,
			}

			resp, err := client.Request("POST", fmt.Sprintf("/api/admin/backups/settings/%s", key), body)
			if err != nil {
				return err
			}

			fmt.Println(resp.Message)
			return nil
		},
	}

	backupCmd.AddCommand(listCmd)
	backupCmd.AddCommand(triggerCmd)
	backupCmd.AddCommand(configCmd)
	backupCmd.AddCommand(setConfigCmd)

	return backupCmd
}

func newAdminSettingsCmd(ctx *config.Context) *cobra.Command {
	settingsCmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage global system settings",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all global settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			resp, err := client.Request("GET", "/api/admin/settings", nil)
			if err != nil {
				return err
			}

			var settings []map[string]interface{}
			if err := json.Unmarshal(resp.Data, &settings); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "KEY\tVALUE\tDESCRIPTION")
			for _, s := range settings {
				// Filter out backup settings from the general list
				if category, ok := s["category"].(string); ok && category == "backup" {
					continue
				}
				fmt.Fprintf(w, "%v\t%v\t%v\n", s["key"], s["value"], s["description"])
			}
			w.Flush()
			return nil
		},
	}

	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Update a global setting",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(ctx.ServerURL, ctx.APIKey)
			client.IsAdmin = true

			key := args[0]
			val := args[1]

			body := map[string]interface{}{
				"value": val,
			}

			resp, err := client.Request("POST", fmt.Sprintf("/api/admin/settings/%s", key), body)
			if err != nil {
				return err
			}

			fmt.Println(resp.Message)
			return nil
		},
	}

	settingsCmd.AddCommand(listCmd)
	settingsCmd.AddCommand(setCmd)

	return settingsCmd
}
