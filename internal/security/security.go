package security

import (
	"fmt"
	"os"
	"syscall"
)


// VerifyNotSymlink checks that the given path is not a symlink.
func VerifyNotSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // doesn't exist yet, OK
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("path '%s' is a symlink — refusing to use (possible attack)", path)
	}
	return nil
}

// HardenProcess sets security limits on the current process:
// - Disables core dumps (RLIMIT_CORE=0) to prevent secrets in crash dumps
// - Locks memory (mlockall) to prevent secrets from being swapped to disk
//
// Both operations are best-effort: failures are logged to stderr but do not
// prevent the daemon from starting. Users should be aware that without
// CAP_IPC_LOCK, secrets may be swappable to disk.
func HardenProcess() {
	if err := syscall.Setrlimit(syscall.RLIMIT_CORE, &syscall.Rlimit{Cur: 0, Max: 0}); err != nil {
		fmt.Fprintf(os.Stderr, "secrun: warning: could not disable core dumps: %v\n", err)
	}

	if err := syscall.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE); err != nil {
		fmt.Fprintf(os.Stderr, "secrun: warning: could not lock memory (secrets may be swappable to disk): %v\n", err)
	}
}
