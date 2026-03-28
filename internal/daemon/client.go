package daemon

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

type Client struct {
	socketPath string
	authToken  string
}

func NewClient(socketPath string, authToken string) *Client {
	return &Client{socketPath: socketPath, authToken: authToken}
}

func (c *Client) send(req Request) (Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 2*time.Second)
	if err != nil { return Response{}, fmt.Errorf("connect to daemon: %w", err) }
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	data, err := EncodeRequest(req)
	if err != nil { return Response{}, fmt.Errorf("encode request: %w", err) }
	if _, err := conn.Write(append(data, '\n')); err != nil { return Response{}, fmt.Errorf("write request: %w", err) }

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() { return Response{}, fmt.Errorf("read response: connection closed") }
	resp, err := DecodeResponse(scanner.Bytes())
	if err != nil { return Response{}, fmt.Errorf("decode response: %w", err) }
	return resp, nil
}

func (c *Client) Ping() error {
	resp, err := c.send(Request{Type: ReqPing, Token: c.authToken})
	if err != nil { return err }
	if !resp.OK { return fmt.Errorf("ping failed: %s", resp.Error) }
	return nil
}

func (c *Client) Get(project string) (map[string]string, error) {
	resp, err := c.send(Request{Type: ReqGet, Token: c.authToken, Project: project})
	if err != nil { return nil, err }
	if !resp.OK { return nil, fmt.Errorf("%s", resp.Error) }
	return resp.Secrets, nil
}

func (c *Client) Put(project string, secrets map[string]string) error {
	resp, err := c.send(Request{Type: ReqPut, Token: c.authToken, Project: project, Secrets: secrets})
	if err != nil { return err }
	if !resp.OK { return fmt.Errorf("put failed: %s", resp.Error) }
	return nil
}

func (c *Client) Clear(project string) error {
	resp, err := c.send(Request{Type: ReqClear, Token: c.authToken, Project: project})
	if err != nil { return err }
	if !resp.OK { return fmt.Errorf("clear failed: %s", resp.Error) }
	return nil
}

func (c *Client) IsRunning() bool { return c.Ping() == nil }
