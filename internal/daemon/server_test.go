package daemon

import (
	"path/filepath"
	"testing"
	"time"
)

const testToken = "test-auth-token-abc123"

func TestDaemonPutAndGet(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	srv := NewServer(sockPath, 1*time.Hour, testToken)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)
	client := NewClient(sockPath, testToken)
	err := client.Put("myproject", map[string]string{"KEY": "val"})
	if err != nil { t.Fatalf("put failed: %v", err) }
	secrets, err := client.Get("myproject")
	if err != nil { t.Fatalf("get failed: %v", err) }
	if secrets["KEY"] != "val" { t.Errorf("expected 'val', got '%s'", secrets["KEY"]) }
}

func TestDaemonRejectsInvalidToken(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	srv := NewServer(sockPath, 1*time.Hour, testToken)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)
	badClient := NewClient(sockPath, "wrong-token")
	err := badClient.Put("myproject", map[string]string{"KEY": "val"})
	if err == nil { t.Fatal("expected auth error with wrong token") }
}

func TestDaemonCacheMiss(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	srv := NewServer(sockPath, 1*time.Hour, testToken)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)
	client := NewClient(sockPath, testToken)
	_, err := client.Get("nonexistent")
	if err == nil { t.Fatal("expected cache miss error") }
}

func TestDaemonClear(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	srv := NewServer(sockPath, 1*time.Hour, testToken)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)
	client := NewClient(sockPath, testToken)
	client.Put("myproject", map[string]string{"KEY": "val"})
	client.Clear("myproject")
	_, err := client.Get("myproject")
	if err == nil { t.Fatal("expected cache miss after clear") }
}

func TestDaemonPing(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	srv := NewServer(sockPath, 1*time.Hour, testToken)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)
	client := NewClient(sockPath, testToken)
	if err := client.Ping(); err != nil { t.Fatalf("ping failed: %v", err) }
}

func TestDaemonTTLExpiry(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	srv := NewServer(sockPath, 100*time.Millisecond, testToken)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)
	client := NewClient(sockPath, testToken)
	client.Put("myproject", map[string]string{"KEY": "val"})
	time.Sleep(200 * time.Millisecond)
	_, err := client.Get("myproject")
	if err == nil { t.Fatal("expected cache miss after TTL expiry") }
}
