package session

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const defaultAgentBootWait = 200 * time.Millisecond

// Client talks to the local session agent over a Unix socket.
type Client struct {
	socketPath string
	ttl        time.Duration
}

// NewClient returns a session-agent client.
func NewClient(socketPath string, ttl time.Duration) *Client {
	return &Client{socketPath: socketPath, ttl: ttl}
}

// DefaultSocketPath returns the socket path scoped by vault path.
func DefaultSocketPath(vaultPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(vaultPath))
	name := "session-" + hex.EncodeToString(sum[:8]) + ".sock"
	return filepath.Join(home, ".ssh", "sshcloak", "run", name), nil
}

// EnsureAgent makes sure the local session agent is reachable.
func (c *Client) EnsureAgent() error {
	if _, err := c.send(request{Action: "ping"}); err == nil {
		return nil
	}
	if err := c.startAgent(); err != nil {
		return err
	}
	time.Sleep(defaultAgentBootWait)
	_, err := c.send(request{Action: "ping"})
	return err
}

// Status reports whether a non-expired cached unlock entry exists.
func (c *Client) Status(sessionID string) (bool, error) {
	resp, err := c.send(request{Action: "status", SessionID: sessionID})
	if err != nil {
		return false, err
	}
	return resp.Found, nil
}

// Get returns cached passphrase for one session id, when present.
func (c *Client) Get(sessionID string) (string, bool, error) {
	resp, err := c.send(request{Action: "get", SessionID: sessionID})
	if err != nil {
		return "", false, err
	}
	if !resp.Found {
		return "", false, nil
	}
	return resp.Value, true, nil
}

// Set stores passphrase for one session id with agent TTL.
func (c *Client) Set(sessionID, passphrase string) error {
	_, err := c.send(request{Action: "set", SessionID: sessionID, Passphrase: passphrase})
	return err
}

// Lock removes one session id entry.
func (c *Client) Lock(sessionID string) error {
	_, err := c.send(request{Action: "lock", SessionID: sessionID})
	return err
}

// Clear drops all cached session entries.
func (c *Client) Clear() error {
	_, err := c.send(request{Action: "clear"})
	return err
}

func (c *Client) send(req request) (response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 500*time.Millisecond)
	if err != nil {
		return response{}, err
	}
	defer conn.Close() //nolint:errcheck

	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return response{}, err
	}
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	if err := enc.Encode(req); err != nil {
		return response{}, err
	}
	var resp response
	if err := dec.Decode(&resp); err != nil {
		return response{}, err
	}
	if resp.Error != "" {
		return response{}, errors.New(resp.Error)
	}
	return resp, nil
}

func (c *Client) startAgent() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exePath, "__session-agent", "--socket", c.socketPath, "--ttl", c.ttl.String())
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("session agent start: %w", err)
	}
	return nil
}
