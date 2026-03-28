package commands

import (
	"fmt"
	"os"

	"github.com/MartSurv/secrun/internal/validate"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(flagProject *string, flagStore *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [project] <KEY>",
		Short: "Remove a secret",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectArg, key string
			if len(args) == 2 {
				projectArg = args[0]
				key = args[1]
			} else {
				key = args[0]
			}
			if err := validate.KeyName(key); err != nil {
				return err
			}
			project, backend, err := resolveContext(flagProject, flagStore, projectArg)
			if err != nil {
				return err
			}
			if err := backend.Delete(project, key); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Deleted %s from project '%s'\n", key, project)
			return nil
		},
	}
}
