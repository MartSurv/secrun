// internal/vault/vault.go
package vault

// Backend is the interface for secret storage backends.
type Backend interface {
	Init(project string) error
	Exists(project string) bool
	Set(project, key, value string) error
	Get(project, key string) (string, error)
	GetAll(project string) (map[string]string, error)
	List(project string) ([]string, error)
	Delete(project, key string) error
	Projects() ([]string, error)
	Count(project string) (int, error)
	BackendName() string
}
