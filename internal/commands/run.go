package commands

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/MartSurv/secrun/internal/config"
	"github.com/MartSurv/secrun/internal/daemon"
	secexec "github.com/MartSurv/secrun/internal/exec"
	"github.com/spf13/cobra"
)

func NewRunCmd(flagProject *string, flagTTL *string, flagNoCache *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [project] -- <command>",
		Short: "Run a command with secrets injected as env vars",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectArg string
			var cmdArgs []string

			dashAt := cmd.ArgsLenAtDash()
			if dashAt == -1 {
				cmdArgs = args
			} else if dashAt == 0 {
				cmdArgs = args
			} else {
				projectArg = args[0]
				cmdArgs = args[dashAt:]
			}

			if len(cmdArgs) == 0 {
				return fmt.Errorf("usage: secrun run [project] -- <command>")
			}

			project, err := resolveProjectName(flagProject, projectArg)
			if err != nil {
				return err
			}

			ttl, err := time.ParseDuration(*flagTTL)
			if err != nil {
				return fmt.Errorf("invalid TTL '%s': %w", *flagTTL, err)
			}

			var secrets map[string]string
			var cachedToken string

			if !*flagNoCache {
				if tokenBytes, err := os.ReadFile(daemon.TokenPath()); err == nil {
					cachedToken = string(tokenBytes)
					client := daemon.NewClient(daemon.SocketPath(), cachedToken)
					if client.IsRunning() {
						if secrets, err = client.Get(project); err == nil {
							return secexec.Run(cmdArgs, secrets)
						}
					}
				}
			}

			backend := fileBackendForProject(project)
			secrets, err = backend.GetAll(project)
			if err != nil {
				return err
			}

			if !*flagNoCache {
				cachedToken, err = startDaemon(ttl)
				if err == nil {
					time.Sleep(50 * time.Millisecond)
					client := daemon.NewClient(daemon.SocketPath(), cachedToken)
					_ = client.Put(project, secrets)
				}
			}

			return secexec.Run(cmdArgs, secrets)
		},
	}
	return cmd
}

const authTokenLen = 32 // 256-bit auth token

func startDaemon(ttl time.Duration) (string, error) {
	tokenBytes := make([]byte, authTokenLen)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	if err := os.MkdirAll(config.ConfigDir(), 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(daemon.TokenPath(), []byte(token), 0600); err != nil {
		return "", fmt.Errorf("write token: %w", err)
	}
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("find executable: %w", err)
	}
	cmd := exec.Command(exePath, "__daemon", "--ttl", ttl.String())
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = strings.NewReader(token)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start daemon: %w", err)
	}
	return token, nil
}
