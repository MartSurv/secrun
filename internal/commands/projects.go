package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/MartSurv/secrun/internal/config"
	"github.com/MartSurv/secrun/internal/vault"
	"github.com/spf13/cobra"
)

func NewProjectsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List all projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fb := vault.NewFileBackend(config.VaultDir(), func() (string, error) {
				return "", fmt.Errorf("password not needed for listing")
			})
			projects, _ := fb.Projects()

			if len(projects) == 0 {
				fmt.Fprintln(os.Stderr, "No projects found. Run 'secrun init' to get started.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "PROJECT")
			for _, p := range projects {
				fmt.Fprintln(w, p)
			}
			w.Flush()
			return nil
		},
	}
}
