package commands

import (
	"fmt"
	"os"

	"github.com/MartSurv/secrun/internal/config"
	"github.com/MartSurv/secrun/internal/prompt"
	"github.com/MartSurv/secrun/internal/resolve"
	"github.com/MartSurv/secrun/internal/vault"
	"github.com/spf13/cobra"
)

func NewInitCmd(flagProject *string, flagStore *string) *cobra.Command {
	return &cobra.Command{
		Use:   "init [project]",
		Short: "Create a new project vault",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectArg := ""
			if len(args) > 0 {
				projectArg = args[0]
			}
			dir, _ := os.Getwd()
			project, err := resolve.ProjectName(coalesce(*flagProject, projectArg), dir)
			if err != nil {
				return err
			}
			cfg, _ := config.Load(config.ConfigPath())
			store := cfg.StoreForProject(project, *flagStore)
			backend := getBackendWithConfirm(store)
			if backend.Exists(project) {
				return fmt.Errorf("project '%s' already exists", project)
			}
			if err := backend.Init(project); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Initialized project '%s' (backend: %s)\n", project, store)
			return nil
		},
	}
}

// cachedPasswordFn wraps a prompt function to cache the password.
// Call clearCache() to force re-prompting (e.g., on wrong password).
func cachedPasswordFn(promptFn func() (string, error)) (passwordFn vault.PasswordFunc, clearCache func()) {
	var cached string
	return func() (string, error) {
		if cached != "" {
			return cached, nil
		}
		pw, err := promptFn()
		if err != nil {
			return "", err
		}
		cached = pw
		return cached, nil
	}, func() { cached = "" }
}

func fileBackend() vault.Backend {
	pwFn, clearFn := cachedPasswordFn(func() (string, error) {
		return prompt.Password("Master password")
	})
	fb := vault.NewFileBackend(config.VaultDir(), pwFn)
	fb.SetClearCacheFn(clearFn)
	return fb
}

func fileBackendWithConfirm() vault.Backend {
	pwFn, clearFn := cachedPasswordFn(func() (string, error) {
		return prompt.PasswordConfirm("Master password")
	})
	fb := vault.NewFileBackend(config.VaultDir(), pwFn)
	fb.SetClearCacheFn(clearFn)
	return fb
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// resolveContext resolves the project name and backend from flags and args.
// projectArg is an optional positional argument; flagProject/flagStore are global flags.
func resolveContext(flagProject, flagStore *string, projectArg string) (string, vault.Backend, error) {
	dir, _ := os.Getwd()
	project, err := resolve.ProjectName(coalesce(*flagProject, projectArg), dir)
	if err != nil {
		return "", nil, err
	}
	cfg, _ := config.Load(config.ConfigPath())
	store := cfg.StoreForProject(project, *flagStore)
	return project, getBackend(store), nil
}
