package commands

import (
	"fmt"

	"github.com/MartSurv/secrun/internal/validate"
	"github.com/spf13/cobra"
)

func NewGetCmd(flagProject *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get [project] <KEY>",
		Short: "Print a single secret value",
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
			project, backend, err := resolveContext(flagProject, projectArg)
			if err != nil {
				return err
			}
			val, err := backend.Get(project, key)
			if err != nil {
				return err
			}
			fmt.Println(val)
			return nil
		},
	}
}
