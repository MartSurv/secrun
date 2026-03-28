package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MartSurv/secrun/internal/config"
	"github.com/MartSurv/secrun/internal/prompt"
	"github.com/MartSurv/secrun/internal/resolve"
	"github.com/MartSurv/secrun/internal/vault"
	"github.com/spf13/cobra"
)

func NewImportCmd(flagProject *string, flagStore *string) *cobra.Command {
	var fromFile string
	cmd := &cobra.Command{
		Use:   "import [project]",
		Short: "Bulk import secrets",
		Long:  "Import from a .env file (--from) or prompt for each key in .env.example / .env.local.example",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := os.Getwd()
			cfg, _ := config.Load(config.ConfigPath())
			projectArg := ""
			if len(args) > 0 {
				projectArg = args[0]
			}
			project, err := resolve.ProjectName(coalesce(*flagProject, projectArg), dir)
			if err != nil {
				return err
			}
			store := cfg.StoreForProject(project, *flagStore)
			backend := getBackend(store)
			if !backend.Exists(project) {
				fmt.Fprintf(os.Stderr, "Project '%s' doesn't exist. Initializing...\n", project)
				initBackend := getBackendWithConfirm(store)
				if err := initBackend.Init(project); err != nil {
					return err
				}
			}
			if fromFile != "" {
				return importFromFile(backend, project, fromFile)
			}
			return importFromExample(backend, project, dir)
		},
	}
	cmd.Flags().StringVar(&fromFile, "from", "", "Path to .env file to import")
	return cmd
}

func importFromFile(backend vault.Backend, project, path string) error {
	secrets, err := parseEnvFile(path)
	if err != nil {
		return err
	}
	count := 0
	for k, v := range secrets {
		if err := backend.Set(project, k, v); err != nil {
			return fmt.Errorf("set %s: %w", k, err)
		}
		count++
	}
	fmt.Fprintf(os.Stderr, "Imported %d secrets into project '%s'\n", count, project)
	return nil
}

func importFromExample(backend vault.Backend, project, dir string) error {
	candidates := []string{".env.example", ".env.local.example"}
	var examplePath string
	for _, c := range candidates {
		p := filepath.Join(dir, c)
		if _, err := os.Stat(p); err == nil {
			examplePath = p
			break
		}
	}
	if examplePath == "" {
		return fmt.Errorf("no .env.example or .env.local.example found in %s", dir)
	}
	fmt.Fprintf(os.Stderr, "Reading keys from %s\n", filepath.Base(examplePath))
	keys, err := parseEnvFileKeys(examplePath)
	if err != nil {
		return err
	}
	count := 0
	for _, key := range keys {
		val, err := prompt.Value(key)
		if err != nil {
			return err
		}
		if val == "" {
			fmt.Fprintf(os.Stderr, "  Skipping %s (empty)\n", key)
			continue
		}
		if err := backend.Set(project, key, val); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
		count++
	}
	fmt.Fprintf(os.Stderr, "Imported %d secrets into project '%s'\n", count, project)
	return nil
}

// scanEnvLines reads an env file and calls fn for each key=value line.
func scanEnvLines(path string, fn func(key, value string)) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		if key != "" {
			fn(key, value)
		}
	}
	return scanner.Err()
}

func parseEnvFile(path string) (map[string]string, error) {
	secrets := make(map[string]string)
	err := scanEnvLines(path, func(key, value string) {
		if value != "" {
			secrets[key] = value
		}
	})
	return secrets, err
}

func parseEnvFileKeys(path string) ([]string, error) {
	var keys []string
	err := scanEnvLines(path, func(key, _ string) {
		keys = append(keys, key)
	})
	return keys, err
}
