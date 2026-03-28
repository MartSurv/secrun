package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyNotSymlinkOnRegularFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "regular.txt")
	os.WriteFile(path, []byte("test"), 0600)

	if err := VerifyNotSymlink(path); err != nil {
		t.Errorf("regular file should pass: %v", err)
	}
}

func TestVerifyNotSymlinkOnSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	os.WriteFile(target, []byte("test"), 0600)
	os.Symlink(target, link)

	if err := VerifyNotSymlink(link); err == nil {
		t.Error("symlink should fail verification")
	}
}

func TestVerifyNotSymlinkOnNonexistent(t *testing.T) {
	if err := VerifyNotSymlink("/nonexistent/path"); err != nil {
		t.Errorf("nonexistent path should pass: %v", err)
	}
}
