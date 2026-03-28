//go:build !darwin

package commands

import "github.com/MartSurv/secrun/internal/vault"

// offerKeychainSave is a no-op on non-macOS platforms.
func offerKeychainSave(project string) {}

// keychainPasswordFn is not available on non-macOS, returns nil.
func keychainPasswordFn(project string) vault.PasswordFunc { return nil }
