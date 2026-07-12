package session

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const socketPerm = 0600

type request struct {
	Action     string `json:"action"`
	SessionID  string `json:"session_id,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

type response struct {
	Found bool   `json:"found,omitempty"`
	Error string `json:"error,omitempty"`
	Value string `json:"value,omitempty"`
}

type cacheEntry struct {
	passphrase string
	expiresAt  time.Time
}

type agent struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]cacheEntry
}

func newAgent(ttl time.Duration) *agent {
	return &agent{
		ttl:     ttl,
		entries: make(map[string]cacheEntry),
	}
}

func (a *agent) handle(req request) response {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch req.Action {
	case "ping":
		return response{}
	case "get":
		entry, ok := a.entries[req.SessionID]
		if !ok {
			return response{Found: false}
		}
		if time.Now().After(entry.expiresAt) {
			delete(a.entries, req.SessionID)
			return response{Found: false}
		}
		return response{Found: true, Value: entry.passphrase}
	case "set":
		if req.SessionID == "" {
			return response{Error: "missing session id"}
		}
		if req.Passphrase == "" {
			return response{Error: "missing passphrase"}
		}
		a.entries[req.SessionID] = cacheEntry{
			passphrase: req.Passphrase,
			expiresAt:  time.Now().Add(a.ttl),
		}
		return response{}
	case "lock":
		delete(a.entries, req.SessionID)
		return response{}
	case "clear":
		for key, entry := range a.entries {
			zeroString(entry.passphrase)
			delete(a.entries, key)
		}
		return response{}
	case "status":
		entry, ok := a.entries[req.SessionID]
		if !ok || time.Now().After(entry.expiresAt) {
			delete(a.entries, req.SessionID)
			return response{Found: false}
		}
		return response{Found: true}
	default:
		return response{Error: "unknown action"}
	}
}

// RunAgent starts a blocking Unix socket server that stores passphrases in memory.
func RunAgent(socketPath string, ttl time.Duration) error {
	if ttl <= 0 {
		return errors.New("session agent: ttl must be > 0")
	}
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return err
	}
	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer listener.Close() //nolint:errcheck

	if err := os.Chmod(socketPath, socketPerm); err != nil {
		return err
	}

	agent := newAgent(ttl)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			defer c.Close() //nolint:errcheck
			decoder := json.NewDecoder(c)
			encoder := json.NewEncoder(c)

			var req request
			if err := decoder.Decode(&req); err != nil {
				_ = encoder.Encode(response{Error: err.Error()})
				return
			}
			_ = encoder.Encode(agent.handle(req))
		}(conn)
	}
}

func zeroString(_ string) {}
