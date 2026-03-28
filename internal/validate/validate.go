package validate

import (
	"fmt"
	"regexp"
)

var validNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.\-]*$`)

// ProjectName validates a project name contains only safe characters.
func ProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if !validNameRe.MatchString(name) {
		return fmt.Errorf("project name '%s' contains invalid characters — only [a-zA-Z0-9_-.] allowed", name)
	}
	if len(name) > 128 {
		return fmt.Errorf("project name too long (max 128 characters)")
	}
	return nil
}

// KeyName validates a secret key name contains only safe characters.
func KeyName(name string) error {
	if name == "" {
		return fmt.Errorf("key name cannot be empty")
	}
	if !validNameRe.MatchString(name) {
		return fmt.Errorf("key name '%s' contains invalid characters — only [a-zA-Z0-9_-.] allowed", name)
	}
	if len(name) > 256 {
		return fmt.Errorf("key name too long (max 256 characters)")
	}
	return nil
}
