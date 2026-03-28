package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func NewExportCmd(flagProject *string) *cobra.Command {
	return &cobra.Command{
		Use:   "export [project]",
		Short: "Print all secrets as KEY=VALUE",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectArg := ""
			if len(args) > 0 {
				projectArg = args[0]
			}
			project, backend, err := resolveContext(flagProject, projectArg)
			if err != nil {
				return err
			}
			secrets, err := backend.GetAll(project)
			if err != nil {
				return err
			}
			keys := make([]string, 0, len(secrets))
			for k := range secrets {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				// Quote values containing newlines to prevent injection
				v := secrets[k]
				if strings.ContainsAny(v, "\n\r") {
					v = "\"" + strings.ReplaceAll(strings.ReplaceAll(v, "\\", "\\\\"), "\"", "\\\"") + "\""
				}
				fmt.Printf("%s=%s\n", k, v)
			}
			return nil
		},
	}
}
