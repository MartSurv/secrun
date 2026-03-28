package vault

import (
	"os"
	"strings"
	"testing"
)

func tempVaultDir(t *testing.T) string {
	t.Helper()
	return t.TempDir() + "/vaults"
}

func newTestBackend(t *testing.T, password string) (*FileBackend, string) {
	t.Helper()
	dir := tempVaultDir(t)
	fn := func() (string, error) { return password, nil }
	return NewFileBackend(dir, fn), dir
}

// 1. TestFileBackendInitAndExists — init creates vault, exists returns true
func TestFileBackendInitAndExists(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	if fb.Exists("myproject") {
		t.Fatal("expected project to not exist before Init")
	}
	if err := fb.Init("myproject"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !fb.Exists("myproject") {
		t.Fatal("expected project to exist after Init")
	}
}

// 2. TestFileBackendSetAndGet — set then get returns same value
func TestFileBackendSetAndGet(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	if err := fb.Init("myproject"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := fb.Set("myproject", "API_KEY", "supersecret"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, err := fb.Get("myproject", "API_KEY")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "supersecret" {
		t.Fatalf("expected 'supersecret', got %q", val)
	}
}

// 3. TestFileBackendGetAll — set multiple, getall returns all
func TestFileBackendGetAll(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	if err := fb.Init("proj"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	secrets := map[string]string{
		"KEY_A": "valueA",
		"KEY_B": "valueB",
		"KEY_C": "valueC",
	}
	for k, v := range secrets {
		if err := fb.Set("proj", k, v); err != nil {
			t.Fatalf("Set(%q) failed: %v", k, err)
		}
	}
	all, err := fb.GetAll("proj")
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	for k, v := range secrets {
		if all[k] != v {
			t.Errorf("expected all[%q]=%q, got %q", k, v, all[k])
		}
	}
}

// 4. TestFileBackendList — list returns sorted key names
func TestFileBackendList(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	if err := fb.Init("proj"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	for _, k := range []string{"ZEBRA", "APPLE", "MANGO"} {
		if err := fb.Set("proj", k, "value"); err != nil {
			t.Fatalf("Set(%q) failed: %v", k, err)
		}
	}
	keys, err := fb.List("proj")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	expected := []string{"APPLE", "MANGO", "ZEBRA"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("keys[%d]: expected %q, got %q", i, k, keys[i])
		}
	}
}

// 5. TestFileBackendDelete — delete removes key
func TestFileBackendDelete(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	if err := fb.Init("proj"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := fb.Set("proj", "KEY", "val"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := fb.Delete("proj", "KEY"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err := fb.Get("proj", "KEY")
	if err == nil {
		t.Fatal("expected error getting deleted key, got nil")
	}
}

// 6. TestFileBackendWrongPassword — wrong password fails decryption
func TestFileBackendWrongPassword(t *testing.T) {
	dir := tempVaultDir(t)
	correctFn := func() (string, error) { return "correct-password", nil }
	wrongFn := func() (string, error) { return "wrong-password", nil }

	fb := NewFileBackend(dir, correctFn)
	if err := fb.Init("proj"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := fb.Set("proj", "KEY", "val"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	fbWrong := NewFileBackend(dir, wrongFn)
	_, err := fbWrong.Get("proj", "KEY")
	if err == nil {
		t.Fatal("expected error with wrong password, got nil")
	}
}

// 7. TestFileBackendProjects — multiple inits, projects lists all
func TestFileBackendProjects(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	for _, p := range []string{"alpha", "beta", "gamma"} {
		if err := fb.Init(p); err != nil {
			t.Fatalf("Init(%q) failed: %v", p, err)
		}
	}
	projects, err := fb.Projects()
	if err != nil {
		t.Fatalf("Projects failed: %v", err)
	}
	expected := []string{"alpha", "beta", "gamma"}
	if len(projects) != len(expected) {
		t.Fatalf("expected %d projects, got %d: %v", len(expected), len(projects), projects)
	}
	for i, p := range expected {
		if projects[i] != p {
			t.Errorf("projects[%d]: expected %q, got %q", i, p, projects[i])
		}
	}
}

// 8. TestFileBackendCount — count returns correct number
func TestFileBackendCount(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	if err := fb.Init("proj"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	count, err := fb.Count("proj")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
	for i, k := range []string{"A", "B", "C"} {
		if err := fb.Set("proj", k, "v"); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		count, err = fb.Count("proj")
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if count != i+1 {
			t.Fatalf("expected count %d, got %d", i+1, count)
		}
	}
}

// 9. TestFileBackendNonexistentProject — getall on missing project errors
func TestFileBackendNonexistentProject(t *testing.T) {
	fb, _ := newTestBackend(t, "testpass")
	_, err := fb.GetAll("doesnotexist")
	if err == nil {
		t.Fatal("expected error for nonexistent project, got nil")
	}
}

// 10. TestFileBackendVaultIsEncrypted — raw file doesn't contain plaintext
func TestFileBackendVaultIsEncrypted(t *testing.T) {
	fb, dir := newTestBackend(t, "testpass")
	if err := fb.Init("proj"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := fb.Set("proj", "SECRET_KEY", "my-very-secret-value"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Read the raw vault file
	vaultPath := dir + "/proj.enc"
	data, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("reading vault file: %v", err)
	}

	// The plaintext should not appear in the raw file
	if strings.Contains(string(data), "my-very-secret-value") {
		t.Fatal("plaintext value found in vault file — not encrypted!")
	}
	if strings.Contains(string(data), "SECRET_KEY") {
		t.Fatal("plaintext key found in vault file — not encrypted!")
	}
}
