package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/MartSurv/secrun/internal/security"
	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 3
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	saltLen       = 16
	magicByte0    = 0x53
	magicByte1    = 0x52
	vaultVersion  = 1
	headerLen     = 4
)

// PasswordFunc is an injectable password prompt function.
type PasswordFunc func() (string, error)

// FileBackend is a file-based vault backend using Argon2id + AES-256-GCM.
type FileBackend struct {
	vaultDir       string
	passwordFn     PasswordFunc
	clearCacheFn   func()
}

// NewFileBackend constructs a new FileBackend.
// clearCacheFn is optional — if set, called to invalidate cached passwords on wrong-password retry.
func NewFileBackend(vaultDir string, passwordFn PasswordFunc) *FileBackend {
	return &FileBackend{vaultDir: vaultDir, passwordFn: passwordFn}
}

// SetClearCacheFn sets the function to clear the password cache (used for retry on wrong password).
func (f *FileBackend) SetClearCacheFn(fn func()) {
	f.clearCacheFn = fn
}

func (f *FileBackend) vaultPath(project string) string {
	return filepath.Join(f.vaultDir, project+".enc")
}

func (f *FileBackend) Exists(project string) bool {
	_, err := os.Stat(f.vaultPath(project))
	return err == nil
}

func (f *FileBackend) Init(project string) error {
	if f.Exists(project) {
		return fmt.Errorf("project '%s' already exists", project)
	}
	if err := os.MkdirAll(f.vaultDir, 0700); err != nil {
		return fmt.Errorf("create vault directory: %w", err)
	}
	return f.writeVault(project, map[string]string{})
}

func (f *FileBackend) Set(project, key, value string) error {
	secrets, err := f.readVault(project)
	if err != nil {
		return err
	}
	secrets[key] = value
	return f.writeVault(project, secrets)
}

func (f *FileBackend) Get(project, key string) (string, error) {
	secrets, err := f.readVault(project)
	if err != nil {
		return "", err
	}
	val, ok := secrets[key]
	if !ok {
		return "", fmt.Errorf("secret '%s' not found in project '%s'", key, project)
	}
	return val, nil
}

func (f *FileBackend) GetAll(project string) (map[string]string, error) {
	return f.readVault(project)
}

func (f *FileBackend) List(project string) ([]string, error) {
	secrets, err := f.readVault(project)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func (f *FileBackend) Delete(project, key string) error {
	secrets, err := f.readVault(project)
	if err != nil {
		return err
	}
	delete(secrets, key)
	return f.writeVault(project, secrets)
}

func (f *FileBackend) Projects() ([]string, error) {
	entries, err := os.ReadDir(f.vaultDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var projects []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".enc") {
			projects = append(projects, strings.TrimSuffix(e.Name(), ".enc"))
		}
	}
	sort.Strings(projects)
	return projects, nil
}

func (f *FileBackend) Count(project string) (int, error) {
	secrets, err := f.readVault(project)
	if err != nil {
		return 0, err
	}
	return len(secrets), nil
}

// deriveKey uses Argon2id to derive a 32-byte key from the password and salt.
func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
}

// readVault decrypts and reads the vault for the given project.
func (f *FileBackend) readVault(project string) (map[string]string, error) {
	path := f.vaultPath(project)

	if err := security.VerifyNotSymlink(path); err != nil {
		return nil, err
	}

	fh, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project '%s' does not exist", project)
		}
		return nil, fmt.Errorf("open vault file: %w", err)
	}
	defer fh.Close()

	// Shared lock for reading
	if err := syscall.Flock(int(fh.Fd()), syscall.LOCK_SH); err != nil {
		return nil, fmt.Errorf("lock vault file: %w", err)
	}
	defer syscall.Flock(int(fh.Fd()), syscall.LOCK_UN) //nolint:errcheck

	data, err := io.ReadAll(fh)
	if err != nil {
		return nil, fmt.Errorf("read vault file: %w", err)
	}

	// Verify magic header: 4 bytes = [0x53, 0x52, version, reserved]
	if len(data) < headerLen+saltLen+12 {
		return nil, fmt.Errorf("vault file too short")
	}
	if data[0] != magicByte0 || data[1] != magicByte1 {
		return nil, fmt.Errorf("invalid vault file magic bytes")
	}
	if data[2] != vaultVersion {
		return nil, fmt.Errorf("unsupported vault version: %d", data[2])
	}
	// data[3] is reserved

	offset := headerLen
	salt := data[offset : offset+saltLen]
	offset += saltLen

	nonce := data[offset : offset+12]
	offset += 12

	ciphertext := data[offset:]

	// AAD: header + salt — ensures these can't be tampered with independently
	aad := data[:headerLen+saltLen]

	// Retry loop: up to 3 attempts for wrong password
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		password, err := f.passwordFn()
		if err != nil {
			return nil, fmt.Errorf("get password: %w", err)
		}

		key := deriveKey(password, salt)

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("create cipher: %w", err)
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("create GCM: %w", err)
		}

		plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
		if err != nil {
			if attempt < maxAttempts {
				fmt.Fprintf(os.Stderr, "Invalid password (%d/%d). Try again.\n", attempt, maxAttempts)
				if f.clearCacheFn != nil {
					f.clearCacheFn()
				}
				continue
			}
			return nil, fmt.Errorf("invalid password (3 attempts failed)")
		}

		var secrets map[string]string
		if err := json.Unmarshal(plaintext, &secrets); err != nil {
			return nil, fmt.Errorf("unmarshal vault: %w", err)
		}
		return secrets, nil
	}

	return nil, fmt.Errorf("invalid password")
}

// writeVault encrypts and writes the vault for the given project atomically.
func (f *FileBackend) writeVault(project string, secrets map[string]string) error {
	plaintext, err := json.Marshal(secrets)
	if err != nil {
		return fmt.Errorf("marshal secrets: %w", err)
	}

	password, err := f.passwordFn()
	if err != nil {
		return fmt.Errorf("get password: %w", err)
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}

	header := []byte{magicByte0, magicByte1, vaultVersion, 0x00}
	aad := make([]byte, 0, headerLen+saltLen)
	aad = append(aad, header...)
	aad = append(aad, salt...)

	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)

	output := make([]byte, 0, headerLen+saltLen+len(nonce)+len(ciphertext))
	output = append(output, header...)
	output = append(output, salt...)
	output = append(output, nonce...)
	output = append(output, ciphertext...)

	if err := os.MkdirAll(f.vaultDir, 0700); err != nil {
		return fmt.Errorf("create vault directory: %w", err)
	}

	vaultPath := f.vaultPath(project)
	tmpPath := vaultPath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open temp file: %w", err)
	}

	if err := syscall.Flock(int(tmpFile.Fd()), syscall.LOCK_EX); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("lock temp file: %w", err)
	}

	_, writeErr := tmpFile.Write(output)
	syncErr := tmpFile.Sync()
	unlockErr := syscall.Flock(int(tmpFile.Fd()), syscall.LOCK_UN)
	closeErr := tmpFile.Close()

	if writeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("write vault: %w", writeErr)
	}
	if syncErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("sync vault: %w", syncErr)
	}
	if unlockErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("unlock temp file: %w", unlockErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", closeErr)
	}

	if err := os.Rename(tmpPath, vaultPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename vault: %w", err)
	}

	if err := os.Chmod(vaultPath, 0600); err != nil {
		return fmt.Errorf("chmod vault: %w", err)
	}

	return nil
}
