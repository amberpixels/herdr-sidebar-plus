// Package herdr is a minimal client for the Herdr socket API: one
// newline-delimited JSON request and response per connection over the Unix
// socket in HERDR_SOCKET_PATH.
package herdr

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// Client talks to the server. It opens a fresh connection per request because
// the server closes each connection after answering.
type Client struct {
	path   string
	nextID int
}

// Dial checks for the socket named by HERDR_SOCKET_PATH, which Herdr injects
// into every plugin command.
func Dial() (*Client, error) {
	path := os.Getenv("HERDR_SOCKET_PATH")
	if path == "" {
		return nil, errors.New("HERDR_SOCKET_PATH is not set; run this from a Herdr plugin")
	}
	return &Client{path: path}, nil
}

type request struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

type response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *APIError       `json:"error"`
}

// APIError is an error response from the server.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string { return e.Code + ": " + e.Message }

// Call sends one request and decodes its result into out (which may be nil).
func (c *Client) Call(method string, params, out any) error {
	if params == nil {
		params = struct{}{}
	}
	c.nextID++
	id := "sidebar-plus:" + strconv.Itoa(c.nextID)
	line, err := json.Marshal(request{ID: id, Method: method, Params: params})
	if err != nil {
		return err
	}
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "unix", c.path)
	if err != nil {
		return fmt.Errorf("connect to herdr: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	if _, err := conn.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	raw, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	var resp response
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("%s: decode response: %w", method, err)
	}
	if resp.Error != nil {
		return fmt.Errorf("%s: %w", method, resp.Error)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(resp.Result, out)
}

// Worktree is the Git provenance Herdr records for worktree-backed workspaces.
type Worktree struct {
	RepoName string `json:"repo_name"`
	RepoRoot string `json:"repo_root"`
}

type Workspace struct {
	ID       string            `json:"workspace_id"`
	Label    string            `json:"label"`
	Tokens   map[string]string `json:"tokens"`
	Worktree *Worktree         `json:"worktree"`
}

// Agent is one live agent, in the order Herdr's agent panel shows them.
type Agent struct {
	PaneID      string            `json:"pane_id"`
	WorkspaceID string            `json:"workspace_id"`
	Tokens      map[string]string `json:"tokens"`
}

func (c *Client) Workspaces() ([]Workspace, error) {
	var out struct {
		Workspaces []Workspace `json:"workspaces"`
	}
	err := c.Call("workspace.list", nil, &out)
	return out.Workspaces, err
}

func (c *Client) Agents() ([]Agent, error) {
	var out struct {
		Agents []Agent `json:"agents"`
	}
	err := c.Call("agent.list", nil, &out)
	return out.Agents, err
}

// ReportWorkspaceTokens patches workspace tokens; a nil value clears the key.
func (c *Client) ReportWorkspaceTokens(workspaceID, source string, tokens map[string]*string) error {
	return c.Call("workspace.report_metadata", map[string]any{
		"workspace_id": workspaceID,
		"source":       source,
		"tokens":       tokens,
	}, nil)
}

// ReportPaneTokens patches pane tokens; a nil value clears the key.
func (c *Client) ReportPaneTokens(paneID, source string, tokens map[string]*string) error {
	return c.Call("pane.report_metadata", map[string]any{
		"pane_id": paneID,
		"source":  source,
		"tokens":  tokens,
	}, nil)
}

func (c *Client) FocusAgent(target string) error {
	return c.Call("agent.focus", map[string]string{"target": target}, nil)
}

func (c *Client) OpenPluginPane(pluginID, entrypoint string) error {
	return c.Call("plugin.pane.open", map[string]any{
		"plugin_id":  pluginID,
		"entrypoint": entrypoint,
	}, nil)
}
