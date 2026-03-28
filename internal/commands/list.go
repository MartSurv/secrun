package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewListCmd(flagProject *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list [project]",
		Short: "List secret names (not values)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectArg := ""
			if len(args) == 1 {
				projectArg = args[0]
			}
			project, backend, err := resolveContext(flagProject, projectArg)
			if err != nil {
				return err
			}
			keys, err := backend.List(project)
			if err != nil {
				return err
			}
			for _, k := range keys {
				fmt.Println(k)
			}
			return nil
		},
	}
}
