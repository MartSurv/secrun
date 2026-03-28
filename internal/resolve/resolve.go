package resolve

import (
	"fmt"
	"path/filepath"

	"github.com/MartSurv/secrun/internal/validate"
)

// ProjectName resolves the project name from the explicit flag or directory path.
// flagValue is the --project flag value (may be empty).
// dir is the current working directory.
func ProjectName(flagValue string, dir string) (string, error) {
	if flagValue != "" {
		if err := validate.ProjectName(flagValue); err != nil {
			return "", err
		}
		return flagValue, nil
	}

	name := filepath.Base(dir)
	if name == "/" || name == "." || name == "" {
		return "", fmt.Errorf("cannot infer project name from directory '%s' — use --project <name>", dir)
	}

	if err := validate.ProjectName(name); err != nil {
		return "", fmt.Errorf("directory name '%s' is not a valid project name: %w — use --project <name>", name, err)
	}

	return name, nil
}
