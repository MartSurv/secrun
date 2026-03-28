//go:build darwin

package vault

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

const markerKey = "__secrun_marker__"

type KeychainBackend struct {
	servicePrefix string
}

func NewKeychainBackend() *KeychainBackend {
	return &KeychainBackend{servicePrefix: "secrun"}
}

func (k *KeychainBackend) BackendName() string { return "keychain" }

func (k *KeychainBackend) serviceName(project string) string {
	return k.servicePrefix + "/" + project
}

func (k *KeychainBackend) Init(project string) error {
	return k.setKeychain(k.serviceName(project), markerKey, "1")
}

func (k *KeychainBackend) Exists(project string) bool {
	_, err := k.getKeychain(k.serviceName(project), markerKey)
	return err == nil
}

func (k *KeychainBackend) Set(project, key, value string) error {
	if !k.Exists(project) { k.Init(project) }
	return k.setKeychain(k.serviceName(project), key, value)
}

func (k *KeychainBackend) Get(project, key string) (string, error) {
	val, err := k.getKeychain(k.serviceName(project), key)
	if err != nil { return "", fmt.Errorf("secret '%s' not found in project '%s'", key, project) }
	return val, nil
}

func (k *KeychainBackend) GetAll(project string) (map[string]string, error) {
	keys, err := k.List(project)
	if err != nil { return nil, err }
	secrets := make(map[string]string, len(keys))
	for _, key := range keys {
		val, err := k.getKeychain(k.serviceName(project), key)
		if err != nil { return nil, fmt.Errorf("read '%s': %w", key, err) }
		secrets[key] = val
	}
	return secrets, nil
}

func (k *KeychainBackend) List(project string) ([]string, error) {
	service := k.serviceName(project)
	out, err := exec.Command("security", "dump-keychain").Output()
	if err != nil { return nil, fmt.Errorf("dump keychain: %w", err) }
	var keys []string
	lines := strings.Split(string(out), "\n")
	inMatchingEntry := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, fmt.Sprintf(`"svce"<blob>="%s"`, service)) { inMatchingEntry = true }
		if inMatchingEntry && strings.Contains(trimmed, `"acct"<blob>="`) {
			acct := extractQuotedValue(trimmed, `"acct"<blob>="`)
			if acct != markerKey { keys = append(keys, acct) }
			inMatchingEntry = false
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func (k *KeychainBackend) Delete(project, key string) error {
	cmd := exec.Command("security", "delete-generic-password", "-s", k.serviceName(project), "-a", key)
	if err := cmd.Run(); err != nil { return fmt.Errorf("delete '%s' from keychain: %w", key, err) }
	return nil
}

func (k *KeychainBackend) Projects() ([]string, error) {
	out, err := exec.Command("security", "dump-keychain").Output()
	if err != nil { return nil, fmt.Errorf("dump keychain: %w", err) }
	seen := map[string]bool{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, fmt.Sprintf(`"svce"<blob>="%s/`, k.servicePrefix)) {
			svc := extractQuotedValue(trimmed, `"svce"<blob>="`)
			project := strings.TrimPrefix(svc, k.servicePrefix+"/")
			if project != "" { seen[project] = true }
		}
	}
	var projects []string
	for p := range seen { projects = append(projects, p) }
	sort.Strings(projects)
	return projects, nil
}

func (k *KeychainBackend) Count(project string) (int, error) {
	keys, err := k.List(project)
	if err != nil { return 0, err }
	return len(keys), nil
}

func (k *KeychainBackend) setKeychain(service, account, password string) error {
	exec.Command("security", "delete-generic-password", "-s", service, "-a", account).Run()
	cmd := exec.Command("security", "add-generic-password", "-s", service, "-a", account, "-w")
	cmd.Stdin = strings.NewReader(password)
	if err := cmd.Run(); err != nil { return fmt.Errorf("store in keychain: %w", err) }
	return nil
}

func (k *KeychainBackend) getKeychain(service, account string) (string, error) {
	out, err := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w").Output()
	if err != nil { return "", err }
	return strings.TrimSpace(string(out)), nil
}

func extractQuotedValue(line, prefix string) string {
	idx := strings.Index(line, prefix)
	if idx == -1 { return "" }
	start := idx + len(prefix)
	end := strings.Index(line[start:], `"`)
	if end == -1 { return "" }
	return line[start : start+end]
}
