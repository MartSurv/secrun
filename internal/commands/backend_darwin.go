//go:build darwin

package commands

import (
	"github.com/MartSurv/secrun/internal/vault"
)

func getBackend(store string) vault.Backend {
	if store == "keychain" {
		return vault.NewKeychainBackend()
	}
	return fileBackend()
}

func getBackendWithConfirm(store string) vault.Backend {
	if store == "keychain" {
		return vault.NewKeychainBackend()
	}
	return fileBackendWithConfirm()
}
