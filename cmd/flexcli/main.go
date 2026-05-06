package main

import (
	"fmt"
	"os"

	"github.com/f1dot4/flexcli/internal/commands"
	"github.com/f1dot4/flexcli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	serverURL   string
	apiKey      string
	contextName string
	rootCfg     *config.Config
	resolvedCtx config.Context
	Version     = "0.2.31"
)

var rootCmd = &cobra.Command{
	Use:     "flexcli",
	Short:   "FlexCLI - FlexCoach Command Line Interface",
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 1. Load config
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
		rootCfg = cfg

		// 2. Resolve target context
		name := contextName
		if name == "" {
			name = rootCfg.CurrentContext
		}

		if name != "" {
			if ctx, ok := rootCfg.Contexts[name]; ok {
				resolvedCtx = ctx
			}
		}

		// 3. Apply overrides (Flag > Env > Context)
		if serverURL != "" {
			resolvedCtx.ServerURL = serverURL
		}

		if apiKey != "" {
			resolvedCtx.APIKey = apiKey
		} else if envKey := os.Getenv("FLEXCLI_API_KEY"); envKey != "" {
			resolvedCtx.APIKey = envKey
		}

		// Validation for non-config commands
		if cmd.Name() != "config" && cmd.Name() != "context" && cmd.Name() != "help" {
			if resolvedCtx.ServerURL == "" {
				return fmt.Errorf("no server URL configured. Run 'config' or use --server")
			}
			if resolvedCtx.APIKey == "" {
				return fmt.Errorf("no API key configured. Run 'config' or use FLEXCLI_API_KEY or --key")
			}
		}

		return nil
	},
}

func main() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flexcli.json)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "FlexCoach server URL override")
	rootCmd.PersistentFlags().StringVar(&apiKey, "key", "", "FlexCoach API key override")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "Use specific context from config")

	rootCmd.AddCommand(commands.NewConfigCmd(&cfgFile, &rootCfg))
	rootCmd.AddCommand(commands.NewContextCmd(&cfgFile, &rootCfg))
	rootCmd.AddCommand(commands.NewProfileCmd(&rootCfg, &resolvedCtx))
	rootCmd.AddCommand(commands.NewConnectCmd(&rootCfg, &resolvedCtx))
	rootCmd.AddCommand(commands.NewPlanCmd(&rootCfg, &resolvedCtx))
	rootCmd.AddCommand(commands.NewAdminCmd(&rootCfg, &resolvedCtx))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
