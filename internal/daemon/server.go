package daemon

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MartSurv/secrun/internal/security"
)

type cacheEntry struct {
	secrets   map[string]string
	expiresAt time.Time
}

type Server struct {
	socketPath string
	ttl        time.Duration
	authToken  string
	cache      map[string]*cacheEntry
	mu         sync.RWMutex
	listener   net.Listener
	done       chan struct{}
	closeOnce  sync.Once
}

func NewServer(socketPath string, ttl time.Duration, authToken string) *Server {
	return &Server{
		socketPath: socketPath, ttl: ttl, authToken: authToken,
		cache: make(map[string]*cacheEntry), done: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	security.HardenProcess()

	dir := filepath.Dir(s.socketPath)
	if err := security.VerifyNotSymlink(dir); err != nil {
		return fmt.Errorf("socket directory: %w", err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil { return fmt.Errorf("create socket directory: %w", err) }

	os.Remove(s.socketPath)
	ln, err := net.Listen("unix", s.socketPath)
	if err != nil { return fmt.Errorf("listen: %w", err) }
	s.listener = ln
	os.Chmod(s.socketPath, 0600)

	go s.cleanupLoop()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.done: return nil
			default: continue
			}
		}
		go s.handleConn(conn)
	}
}

func (s *Server) Stop() {
	s.closeOnce.Do(func() {
		close(s.done)
		if s.listener != nil { s.listener.Close() }
		os.Remove(s.socketPath)
	})
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() { return }
	req, err := DecodeRequest(scanner.Bytes())
	if err != nil { s.writeResponse(conn, Response{OK: false, Error: "invalid request"}); return }

	if req.Token != s.authToken {
		s.writeResponse(conn, Response{OK: false, Error: "unauthorized"})
		return
	}

	switch req.Type {
	case ReqPing:
		s.writeResponse(conn, Response{OK: true})
	case ReqGet:
		s.mu.Lock()
		entry, ok := s.cache[req.Project]
		if !ok || time.Now().After(entry.expiresAt) {
			if ok { delete(s.cache, req.Project) }
			s.mu.Unlock()
			s.writeResponse(conn, Response{OK: false, Error: "cache miss"})
			return
		}
		secrets := entry.secrets
		s.mu.Unlock()
		s.writeResponse(conn, Response{OK: true, Secrets: secrets})
	case ReqPut:
		s.mu.Lock()
		s.cache[req.Project] = &cacheEntry{secrets: req.Secrets, expiresAt: time.Now().Add(s.ttl)}
		s.mu.Unlock()
		s.writeResponse(conn, Response{OK: true})
	case ReqClear:
		s.mu.Lock(); delete(s.cache, req.Project); s.mu.Unlock()
		s.writeResponse(conn, Response{OK: true})
	default:
		s.writeResponse(conn, Response{OK: false, Error: "unknown request type"})
	}
}

func (s *Server) writeResponse(conn net.Conn, resp Response) {
	data, _ := EncodeResponse(resp)
	conn.Write(append(data, '\n'))
}

func (s *Server) cleanupLoop() {
	interval := s.ttl / 4
	if interval < 30*time.Second { interval = 30 * time.Second }
	if interval > 5*time.Minute { interval = 5 * time.Minute }
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done: return
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for project, entry := range s.cache {
				if now.After(entry.expiresAt) { delete(s.cache, project) }
			}
			empty := len(s.cache) == 0
			s.mu.Unlock()
			if empty { s.Stop(); return }
		}
	}
}
