package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/f1dot4/flexcli/internal/api"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

// NewProfileDataCmd builds the profile data command tree, grouping
// manual sync, imported activities, and imported health metrics.
func NewProfileDataCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Sync & data: manual sync, activities, health metrics",
	}

	cmd.AddCommand(newDataSyncCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataActivityCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataHealthMetricCmd(rootCfg, resolvedCtx))

	return cmd
}

// ---------------------------------------------------------------------------
// data sync
// ---------------------------------------------------------------------------

func newDataSyncCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Manually trigger Garmin or Withings synchronization",
	}

	cmd.AddCommand(newDataSyncGarminCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataSyncWithingsCmd(rootCfg, resolvedCtx))

	return cmd
}

func newDataSyncGarminCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "garmin",
		Short: "Sync data from Garmin Connect",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)
			events, err := client.PostSSE("/api/sync/garmin/stream")
			if err != nil {
				return err
			}

			var syncSuccess bool
			var finalError string

			for event := range events {
				if event.Event == "result" {
					var result struct {
						Success bool   `json:"success"`
						Error   string `json:"error"`
					}
					if err := json.Unmarshal([]byte(event.Data), &result); err == nil {
						syncSuccess = result.Success
						finalError = result.Error
					}
				} else {
					fmt.Printf("🔄 %s\n", event.Data)
				}
			}

			if syncSuccess {
				fmt.Println("✅ Garmin synchronization complete.")
			} else {
				if finalError != "" {
					fmt.Printf("❌ Garmin synchronization failed: %s\n", finalError)
				} else {
					fmt.Println("❌ Garmin synchronization failed.")
				}
				return fmt.Errorf("sync failed")
			}

			return nil
		},
	}
}

func newDataSyncWithingsCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "withings",
		Short: "Sync data from Withings",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)
			events, err := client.PostSSE("/api/sync/withings/stream")
			if err != nil {
				return err
			}

			var syncSuccess bool
			var finalError string

			for event := range events {
				if event.Event == "result" {
					var result struct {
						Success bool   `json:"success"`
						Error   string `json:"error"`
					}
					if err := json.Unmarshal([]byte(event.Data), &result); err == nil {
						syncSuccess = result.Success
						finalError = result.Error
					}
				} else {
					fmt.Printf("🔄 %s\n", event.Data)
				}
			}

			if syncSuccess {
				fmt.Println("✅ Withings synchronization complete.")
			} else {
				if finalError != "" {
					fmt.Printf("❌ Withings synchronization failed: %s\n", finalError)
				} else {
					fmt.Println("❌ Withings synchronization failed.")
				}
				return fmt.Errorf("sync failed")
			}

			return nil
		},
	}
}
// ---------------------------------------------------------------------------
// data activity (alias: act)
// ---------------------------------------------------------------------------

func newDataActivityCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activity",
		Aliases: []string{"act"},
		Short:   "Manage Garmin activities (alias: act): list, download, upload, delete",
	}

	cmd.AddCommand(newDataActivityListCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataActivityDownloadCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataActivityUploadCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataActivityDeleteCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataActivityRenameCmd(rootCfg, resolvedCtx))

	return cmd
}

func newDataActivityListCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var page int
	var pageSize int
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List synced activities with their Garmin activity IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			path := fmt.Sprintf("/api/activities?page=%d&page_size=%d", page, pageSize)
			resp, err := client.Request("GET", path, nil)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			var data struct {
				Activities []struct {
					GarminActivityID string   `json:"garmin_activity_id"`
					Type             string   `json:"type"`
					Description      string   `json:"description"`
					StartTime        string   `json:"start_time"`
					DurationMinutes  int      `json:"duration_minutes"`
					DistanceKm       *float64 `json:"distance_km"`
				} `json:"activities"`
				TotalEntries int `json:"total_entries"`
				TotalPages   int `json:"total_pages"`
				CurrentPage  int `json:"current_page"`
			}
			if err := json.Unmarshal(resp.Data, &data); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if len(data.Activities) == 0 {
				fmt.Println("No activities found.")
				return nil
			}

			fmt.Printf("Activities (page %d/%d, %d total):\n\n", data.CurrentPage, data.TotalPages, data.TotalEntries)
			fmt.Printf("  %-16s  %-20s  %-22s  %6s  %s\n", "GARMIN ID", "DATE", "TYPE", "MIN", "DESCRIPTION")
			fmt.Printf("  %-16s  %-20s  %-22s  %6s  %s\n", "────────────────", "────────────────────", "──────────────────────", "──────", "───────────")
			for _, a := range data.Activities {
				dateStr := a.StartTime
				if len(dateStr) > 16 {
					dateStr = dateStr[:16]
				}
				actType := strings.Replace(a.Type, "_", " ", -1)
				desc := a.Description
				fmt.Printf("  %-16s  %-20s  %-22s  %6d  %s\n", a.GarminActivityID, dateStr, actType, a.DurationMinutes, desc)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Number of activities per page")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newDataActivityDownloadCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "download <activity_id>",
		Short: "Download an activity's original FIT file from Garmin Connect",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			activityID := args[0]
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			if output == "" {
				output = activityID + ".zip"
			}

			path := fmt.Sprintf("/api/activity/%s/download", activityID)
			if err := client.DownloadFile(path, output); err != nil {
				return fmt.Errorf("download failed: %w", err)
			}

			fmt.Printf("Downloaded activity %s to %s\n", activityID, output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: <activity_id>.zip)")
	return cmd
}

func newDataActivityUploadCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a FIT/GPX/TCX file to Garmin Connect",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", filePath)
			}
			ext := strings.ToLower(filepath.Ext(filePath))
			if ext != ".fit" && ext != ".gpx" && ext != ".tcx" {
				return fmt.Errorf("unsupported file type %s (must be .fit, .gpx, or .tcx)", ext)
			}

			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)
			resp, err := client.UploadFile("/api/activity/upload", filePath)
			if err != nil {
				return fmt.Errorf("upload failed: %w", err)
			}

			if resp.Success {
				fmt.Printf("Uploaded %s successfully.\n", filepath.Base(filePath))
			} else {
				fmt.Printf("Upload failed: %s\n", resp.Message)
			}
			return nil
		},
	}
}

func newDataActivityDeleteCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <activity_id>",
		Short: "Delete an activity from Garmin Connect (currently disabled)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Activity deletion is currently not implemented due to safety reasons.")
			fmt.Println("Please delete activities manually via Garmin Connect.")
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// data healthmetric (alias: hm)
// ---------------------------------------------------------------------------

func newDataHealthMetricCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "healthmetric",
		Aliases: []string{"hm"},
		Short:   "View imported health metrics (alias: hm): list, show, delete",
	}

	cmd.AddCommand(newDataHealthMetricListCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataHealthMetricShowCmd(rootCfg, resolvedCtx))
	cmd.AddCommand(newDataHealthMetricDeleteCmd(rootCfg, resolvedCtx))

	return cmd
}

func newDataHealthMetricListCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var page int
	var pageSize int
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List imported health metrics (paginated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			path := fmt.Sprintf("/api/healthmetrics?page=%d&page_size=%d", page, pageSize)
			resp, err := client.Request("GET", path, nil)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			var data struct {
				Metrics []struct {
					ID                int      `json:"id"`
					Date              string   `json:"date"`
					Source            string   `json:"source"`
					WeightKg          *float64 `json:"weight_kg"`
					RestingHeartRate  *int     `json:"resting_heart_rate"`
					HRVScore          *float64 `json:"hrv_score"`
					SleepHours        *float64 `json:"sleep_hours"`
					CyclingFTP        *float64 `json:"cycling_ftp"`
					CyclingLTHR       *float64 `json:"cycling_lthr"`
					RunningFTP        *float64 `json:"running_ftp"`
					RunningLTHR       *float64 `json:"running_lthr"`
				} `json:"metrics"`
				TotalEntries int `json:"total_entries"`
				TotalPages   int `json:"total_pages"`
				CurrentPage  int `json:"current_page"`
			}
			if err := json.Unmarshal(resp.Data, &data); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if len(data.Metrics) == 0 {
				fmt.Println("No health metrics found.")
				return nil
			}

			fmt.Printf("Health metrics (page %d/%d, %d total):\n\n", data.CurrentPage, data.TotalPages, data.TotalEntries)
			fmt.Printf("  %-12s  %-10s  %8s  %5s  %5s  %5s  %6s  %6s  %6s  %6s\n",
				"DATE", "SOURCE", "WEIGHT", "RHR", "HRV", "SLEEP", "C-FTP", "C-LTHR", "R-FTP", "R-LTHR")
			fmt.Printf("  %-12s  %-10s  %8s  %5s  %5s  %5s  %6s  %6s  %6s  %6s\n",
				"──────────", "──────────", "────────", "─────", "─────", "─────", "──────", "──────", "──────", "──────")
			for _, m := range data.Metrics {
				weight := "  -"
				if m.WeightKg != nil {
					weight = fmt.Sprintf("%6.1fkg", *m.WeightKg)
				}
				rhr := "  -"
				if m.RestingHeartRate != nil {
					rhr = fmt.Sprintf("%5d", *m.RestingHeartRate)
				}
				hrv := "  -"
				if m.HRVScore != nil {
					hrv = fmt.Sprintf("%5.1f", *m.HRVScore)
				}
				sleep := "  -"
				if m.SleepHours != nil {
					sleep = fmt.Sprintf("%5.1f", *m.SleepHours)
				}
				cftp := "  -"
				if m.CyclingFTP != nil {
					cftp = fmt.Sprintf("%6.0f", *m.CyclingFTP)
				}
				clthr := "  -"
				if m.CyclingLTHR != nil {
					clthr = fmt.Sprintf("%6.0f", *m.CyclingLTHR)
				}
				rftp := "  -"
				if m.RunningFTP != nil {
					rftp = fmt.Sprintf("%6.0f", *m.RunningFTP)
				}
				rlthr := "  -"
				if m.RunningLTHR != nil {
					rlthr = fmt.Sprintf("%6.0f", *m.RunningLTHR)
				}
				fmt.Printf("  %-12s  %-10s  %8s  %5s  %5s  %5s  %6s  %6s  %6s  %6s\n",
					m.Date, m.Source, weight, rhr, hrv, sleep, cftp, clthr, rftp, rlthr)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Number of metrics per page")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newDataHealthMetricShowCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "show <date>",
		Short: "Show aggregated health metric for a specific date (YYYY-MM-DD)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			date := args[0]
			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			resp, err := client.Request("GET", "/api/healthmetric/"+date, nil)
			if err != nil {
				return err
			}

			if asJSON {
				fmt.Println(string(resp.Data))
				return nil
			}

			var metric map[string]interface{}
			if err := json.Unmarshal(resp.Data, &metric); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			fmt.Printf("Health metric for %s\n", date)
			for k, v := range metric {
				if v == nil {
					continue
				}
				fmt.Printf("  • %-28s %v\n", k+":", v)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	return cmd
}

func newDataHealthMetricDeleteCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a health metric (currently disabled)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Health metric deletion is currently not implemented due to safety reasons.")
			fmt.Println("Please remove records manually if absolutely necessary.")
			return nil
		},
	}
}

func newDataActivityRenameCmd(rootCfg **config.Config, resolvedCtx *config.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <activity_id> <title>",
		Short: "Rename an activity in Garmin Connect",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			activityID := args[0]
			title := args[1]

			client := api.NewClient(resolvedCtx.ServerURL, resolvedCtx.APIKey)

			body := map[string]string{
				"name": title,
			}

			path := fmt.Sprintf("/api/activity/%s/name", activityID)
			resp, err := client.Request("PUT", path, body)
			if err != nil {
				return fmt.Errorf("rename failed: %w", err)
			}

			if resp.Success {
				fmt.Printf("✅ Activity %s renamed to: %s\n", activityID, title)
			} else {
				fmt.Printf("❌ Rename failed: %s\n", resp.Message)
			}
			return nil
		},
	}
}
