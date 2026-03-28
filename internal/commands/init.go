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

func NewInitCmd(flagProject *string) *cobra.Command {
	return &cobra.Command{
		Use:   "init [project]",
		Short: "Create a new project vault",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectArg := ""
			if len(args) > 0 {
				projectArg = args[0]
			}
			project, err := resolveProjectName(flagProject, projectArg)
			if err != nil {
				return err
			}

			backend := fileBackendWithConfirm()
			if backend.Exists(project) {
				return fmt.Errorf("project '%s' already exists", project)
			}
			if err := backend.Init(project); err != nil {
				return err
			}

			offerKeychainSave(project)

			fmt.Fprintf(os.Stderr, "Initialized project '%s'\n", project)
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

// fileBackendForProject returns a file backend that tries Keychain first (on macOS).
func fileBackendForProject(project string) vault.Backend {
	kcFn := keychainPasswordFn(project)
	if kcFn != nil {
		fb := vault.NewFileBackend(config.VaultDir(), kcFn)
		return fb
	}
	return fileBackend()
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// resolveProjectName resolves the project name from flag and positional arg.
func resolveProjectName(flagProject *string, projectArg string) (string, error) {
	dir, _ := os.Getwd()
	return resolve.ProjectName(coalesce(*flagProject, projectArg), dir)
}

// resolveContext resolves the project name and returns a Keychain-aware backend.
func resolveContext(flagProject *string, projectArg string) (string, vault.Backend, error) {
	project, err := resolveProjectName(flagProject, projectArg)
	if err != nil {
		return "", nil, err
	}
	return project, fileBackendForProject(project), nil
}
