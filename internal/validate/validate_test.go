package validate

import (
	"testing"
)

func TestValidProjectName(t *testing.T) {
	valid := []string{"my-project", "page-probe", "app_v2", "backend.prod", "ABC123"}
	for _, name := range valid {
		if err := ProjectName(name); err != nil {
			t.Errorf("expected '%s' to be valid, got error: %v", name, err)
		}
	}
}

func TestInvalidProjectName(t *testing.T) {
	invalid := []string{"", "../escape", "my project", "path/traversal", "semi;colon", "back`tick", "$var", "null\x00byte"}
	for _, name := range invalid {
		if err := ProjectName(name); err == nil {
			t.Errorf("expected '%s' to be invalid, got nil", name)
		}
	}
}

func TestValidKeyName(t *testing.T) {
	valid := []string{"API_KEY", "STRIPE_SECRET_KEY", "DB.URL", "my-key", "KEY123"}
	for _, name := range valid {
		if err := KeyName(name); err != nil {
			t.Errorf("expected '%s' to be valid, got error: %v", name, err)
		}
	}
}

func TestInvalidKeyName(t *testing.T) {
	invalid := []string{"", "has space", "semi;colon", "../path", "null\x00byte"}
	for _, name := range invalid {
		if err := KeyName(name); err == nil {
			t.Errorf("expected '%s' to be invalid, got nil", name)
		}
	}
}
