package resolve

import (
	"testing"
)

func TestResolveFromExplicitFlag(t *testing.T) {
	name, err := ProjectName("my-project", "/some/random/dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "my-project" {
		t.Errorf("expected 'my-project', got '%s'", name)
	}
}

func TestResolveFromDirectory(t *testing.T) {
	name, err := ProjectName("", "/home/user/workspace/page-probe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "page-probe" {
		t.Errorf("expected 'page-probe', got '%s'", name)
	}
}

func TestResolveFromRootDirectoryFails(t *testing.T) {
	_, err := ProjectName("", "/")
	if err == nil {
		t.Fatal("expected error for root directory, got nil")
	}
}

func TestResolveFromPositionalArg(t *testing.T) {
	name, err := ProjectName("", "/some/dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "dir" {
		t.Errorf("expected 'dir', got '%s'", name)
	}
}
