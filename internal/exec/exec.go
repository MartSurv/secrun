package exec

import (
	"fmt"
	"os"
	osExec "os/exec"
	"strings"
	"syscall"
)

func BuildEnv(base []string, secrets map[string]string) []string {
	envMap := make(map[string]string, len(base))
	order := make([]string, 0, len(base))
	for _, e := range base {
		idx := strings.IndexByte(e, '=')
		if idx == -1 { continue }
		key := e[:idx]
		envMap[key] = e[idx+1:]
		order = append(order, key)
	}
	for k, v := range secrets {
		if _, exists := envMap[k]; !exists { order = append(order, k) }
		envMap[k] = v
	}
	result := make([]string, 0, len(order))
	for _, k := range order { result = append(result, k+"="+envMap[k]) }
	return result
}

func Run(args []string, secrets map[string]string) error {
	if len(args) == 0 { return fmt.Errorf("no command specified") }
	binary, err := osExec.LookPath(args[0])
	if err != nil { return fmt.Errorf("command not found: %s", args[0]) }
	env := BuildEnv(os.Environ(), secrets)
	return syscall.Exec(binary, args, env)
}
