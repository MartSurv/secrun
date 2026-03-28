package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/MartSurv/secrun/internal/commands"
	"github.com/MartSurv/secrun/internal/daemon"
	"github.com/spf13/cobra"
)

var (
	flagProject string
	flagTTL     string
	flagNoCache bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "secrun",
		Short: "Secure environment variable runner",
		Long:  "Keep secrets out of project directories. Store them encrypted, inject at runtime.",
	}

	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", "Explicit project name (overrides directory inference)")
	rootCmd.PersistentFlags().StringVar(&flagTTL, "ttl", "4h", "Session cache duration")
	rootCmd.PersistentFlags().BoolVar(&flagNoCache, "no-cache", false, "Skip session daemon, prompt every time")

	rootCmd.AddCommand(commands.NewInitCmd(&flagProject))
	rootCmd.AddCommand(commands.NewSetCmd(&flagProject))
	rootCmd.AddCommand(commands.NewGetCmd(&flagProject))
	rootCmd.AddCommand(commands.NewListCmd(&flagProject))
	rootCmd.AddCommand(commands.NewDeleteCmd(&flagProject))
	rootCmd.AddCommand(commands.NewImportCmd(&flagProject))
	rootCmd.AddCommand(commands.NewExportCmd(&flagProject))
	rootCmd.AddCommand(commands.NewProjectsCmd())
	rootCmd.AddCommand(commands.NewRunCmd(&flagProject, &flagTTL, &flagNoCache))
	rootCmd.AddCommand(newDaemonCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newDaemonCmd() *cobra.Command {
	var ttlFlag string
	cmd := &cobra.Command{
		Use:    "__daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ttl, err := time.ParseDuration(ttlFlag)
			if err != nil {
				return err
			}
			tokenBytes, err := io.ReadAll(os.Stdin)
			if err != nil || len(tokenBytes) == 0 {
				return fmt.Errorf("auth token required (via stdin)")
			}
			token := strings.TrimSpace(string(tokenBytes))
			srv := daemon.NewServer(daemon.SocketPath(), ttl, token)
			return srv.Start()
		},
	}
	cmd.Flags().StringVar(&ttlFlag, "ttl", "4h", "TTL duration")
	return cmd
}
