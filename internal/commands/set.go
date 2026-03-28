package commands

import (
	"fmt"
	"os"

	"github.com/MartSurv/secrun/internal/prompt"
	"github.com/MartSurv/secrun/internal/validate"
	"github.com/spf13/cobra"
)

func NewSetCmd(flagProject *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set [project] <KEY> [VALUE]",
		Short: "Store a secret",
		Long:  "Store a secret. If VALUE is omitted, prompts interactively to avoid shell history leaks.",
		Args:  cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var project, key, value string
			switch len(args) {
			case 1:
				// secrun set KEY
				p, err := resolveProjectName(flagProject, "")
				if err != nil {
					return err
				}
				project = p
				key = args[0]
			case 2:
				// Ambiguous: secrun set PROJECT KEY  or  secrun set KEY VALUE
				// Heuristic: if a vault exists for args[0], treat as project
				fb := fileBackendForProject(args[0])
				if fb.Exists(args[0]) {
					project = args[0]
					key = args[1]
				} else {
					p, err := resolveProjectName(flagProject, "")
					if err != nil {
						return err
					}
					project = p
					key = args[0]
					value = args[1]
				}
			case 3:
				// secrun set PROJECT KEY VALUE
				project = args[0]
				key = args[1]
				value = args[2]
			}
			if err := validate.KeyName(key); err != nil {
				return err
			}
			if value == "" {
				v, err := prompt.Value(fmt.Sprintf("Value for %s", key))
				if err != nil {
					return err
				}
				value = v
			} else {
				fmt.Fprintln(os.Stderr, "Warning: value passed as argument and may appear in shell history. Use 'secrun set KEY' (interactive) instead.")
			}
			backend := fileBackendForProject(project)
			if !backend.Exists(project) {
				return fmt.Errorf("project '%s' not found — run 'secrun init %s' first", project, project)
			}
			if err := backend.Set(project, key, value); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Set %s in project '%s'\n", key, project)
			return nil
		},
	}
}
