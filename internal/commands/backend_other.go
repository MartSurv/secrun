//go:build !darwin

package commands

import (
	"fmt"

	"github.com/MartSurv/secrun/internal/vault"
)

func getBackend(store string) vault.Backend {
	if store == "keychain" {
		fmt.Println("Warning: keychain backend is only available on macOS. Using file backend.")
	}
	return fileBackend()
}

func getBackendWithConfirm(store string) vault.Backend {
	if store == "keychain" {
		fmt.Println("Warning: keychain backend is only available on macOS. Using file backend.")
	}
	return fileBackendWithConfirm()
}
