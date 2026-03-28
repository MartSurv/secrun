package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func NewExportCmd(flagProject *string, flagStore *string) *cobra.Command {
	return &cobra.Command{
		Use:   "export [project]",
		Short: "Print all secrets as KEY=VALUE",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectArg := ""
			if len(args) > 0 {
				projectArg = args[0]
			}
			project, backend, err := resolveContext(flagProject, flagStore, projectArg)
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
				fmt.Printf("%s=%s\n", k, secrets[k])
			}
			return nil
		},
	}
}
