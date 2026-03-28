package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MartSurv/secrun/internal/vault"
)

func TestFullFlowFileBackend(t *testing.T) {
	vaultDir := filepath.Join(t.TempDir(), "vaults")
	password := "test-master-password"
	fb := vault.NewFileBackend(vaultDir, func() (string, error) { return password, nil })

	err := fb.Init("test-project")
	if err != nil { t.Fatalf("init: %v", err) }

	secrets := map[string]string{"STRIPE_KEY": "sk_test_123", "DB_URL": "postgres://localhost/mydb", "API_SECRET": "super-secret"}
	for k, v := range secrets {
		if err := fb.Set("test-project", k, v); err != nil { t.Fatalf("set %s: %v", k, err) }
	}

	keys, err := fb.List("test-project")
	if err != nil { t.Fatalf("list: %v", err) }
	if len(keys) != 3 { t.Fatalf("expected 3 keys, got %d", len(keys)) }

	val, err := fb.Get("test-project", "STRIPE_KEY")
	if err != nil { t.Fatalf("get: %v", err) }
	if val != "sk_test_123" { t.Errorf("expected 'sk_test_123', got '%s'", val) }

	all, err := fb.GetAll("test-project")
	if err != nil { t.Fatalf("getall: %v", err) }
	if len(all) != 3 { t.Fatalf("expected 3 secrets, got %d", len(all)) }

	err = fb.Delete("test-project", "API_SECRET")
	if err != nil { t.Fatalf("delete: %v", err) }
	keys, _ = fb.List("test-project")
	if len(keys) != 2 { t.Fatalf("expected 2 keys after delete, got %d", len(keys)) }

	count, err := fb.Count("test-project")
	if err != nil { t.Fatalf("count: %v", err) }
	if count != 2 { t.Errorf("expected count 2, got %d", count) }

	projects, err := fb.Projects()
	if err != nil { t.Fatalf("projects: %v", err) }
	if len(projects) != 1 || projects[0] != "test-project" { t.Errorf("unexpected projects: %v", projects) }

	data, _ := os.ReadFile(filepath.Join(vaultDir, "test-project.enc"))
	if containsBytes(data, []byte("sk_test_123")) { t.Fatal("vault file contains plaintext secret") }
}

func containsBytes(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) { return true }
	}
	return false
}
