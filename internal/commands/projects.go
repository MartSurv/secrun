package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/MartSurv/secrun/internal/config"
	"github.com/MartSurv/secrun/internal/vault"
	"github.com/spf13/cobra"
)

func NewProjectsCmd(flagStore *string) *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List all projects with backend type and secret count",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			type projectInfo struct {
				name, backend string
				count         int
			}
			var projects []projectInfo

			fileBackend := vault.NewFileBackend(config.VaultDir(), func() (string, error) {
				return "", fmt.Errorf("password not needed for listing")
			})
			fileProjects, _ := fileBackend.Projects()
			for _, p := range fileProjects {
				projects = append(projects, projectInfo{name: p, backend: "file", count: -1})
			}

			if len(projects) == 0 {
				fmt.Fprintln(os.Stderr, "No projects found. Run 'secrun init <project>' to get started.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "PROJECT\tBACKEND\tSECRETS")
			for _, p := range projects {
				countStr := fmt.Sprintf("%d", p.count)
				if p.count == -1 {
					countStr = "(locked)"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", p.name, p.backend, countStr)
			}
			w.Flush()
			return nil
		},
	}
}
