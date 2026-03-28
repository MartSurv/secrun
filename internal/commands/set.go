package commands

import (
	"fmt"
	"os"

	"github.com/MartSurv/secrun/internal/config"
	"github.com/MartSurv/secrun/internal/prompt"
	"github.com/MartSurv/secrun/internal/resolve"
	"github.com/MartSurv/secrun/internal/validate"
	"github.com/spf13/cobra"
)

func NewSetCmd(flagProject *string, flagStore *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set [project] <KEY> [VALUE]",
		Short: "Store a secret",
		Long:  "Store a secret. If VALUE is omitted, prompts interactively to avoid shell history leaks.",
		Args:  cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := os.Getwd()
			cfg, _ := config.Load(config.ConfigPath())
			var project, key, value string
			switch len(args) {
			case 1:
				p, err := resolve.ProjectName(*flagProject, dir)
				if err != nil {
					return err
				}
				project = p
				key = args[0]
			case 2:
				store := cfg.StoreForProject(args[0], *flagStore)
				backend := getBackend(store)
				if backend.Exists(args[0]) {
					project = args[0]
					key = args[1]
				} else {
					p, err := resolve.ProjectName(*flagProject, dir)
					if err != nil {
						return err
					}
					project = p
					key = args[0]
					value = args[1]
				}
			case 3:
				project = args[0]
				key = args[1]
				value = args[2]
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
			if err := validate.KeyName(key); err != nil {
				return err
			}
			store := cfg.StoreForProject(project, *flagStore)
			backend := getBackend(store)
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
