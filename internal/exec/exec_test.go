package exec

import "testing"

func TestBuildEnv(t *testing.T) {
	base := []string{"HOME=/home/user", "PATH=/usr/bin", "EXISTING=old"}
	secrets := map[string]string{"API_KEY": "sk-123", "EXISTING": "new"}
	result := BuildEnv(base, secrets)
	found := map[string]string{}
	for _, e := range result {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' { found[e[:i]] = e[i+1:]; break }
		}
	}
	if found["HOME"] != "/home/user" { t.Errorf("HOME not preserved") }
	if found["PATH"] != "/usr/bin" { t.Errorf("PATH not preserved") }
	if found["API_KEY"] != "sk-123" { t.Errorf("API_KEY not injected") }
	if found["EXISTING"] != "new" { t.Errorf("expected secret to override existing: got '%s'", found["EXISTING"]) }
}

func TestBuildEnvEmpty(t *testing.T) {
	base := []string{"HOME=/home/user"}
	secrets := map[string]string{}
	result := BuildEnv(base, secrets)
	if len(result) != 1 { t.Errorf("expected 1 env var, got %d", len(result)) }
}
