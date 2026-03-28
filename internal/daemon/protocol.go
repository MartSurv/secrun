package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type RequestType string

const (
	ReqGet   RequestType = "GET"
	ReqPut   RequestType = "PUT"
	ReqClear RequestType = "CLEAR"
	ReqPing  RequestType = "PING"
)

type Request struct {
	Type    RequestType       `json:"type"`
	Token   string            `json:"token"`
	Project string            `json:"project,omitempty"`
	Secrets map[string]string `json:"secrets,omitempty"`
}

type Response struct {
	OK      bool              `json:"ok"`
	Secrets map[string]string `json:"secrets,omitempty"`
	Error   string            `json:"error,omitempty"`
}

func EncodeRequest(r Request) ([]byte, error)    { return json.Marshal(r) }
func DecodeRequest(data []byte) (Request, error)  { var r Request; err := json.Unmarshal(data, &r); return r, err }
func EncodeResponse(r Response) ([]byte, error)   { return json.Marshal(r) }
func DecodeResponse(data []byte) (Response, error) { var r Response; err := json.Unmarshal(data, &r); return r, err }

func SocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "secrun", "daemon.sock")
}

func TokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "secrun", "daemon.token")
}
