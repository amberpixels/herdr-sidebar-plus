// Command herdr-sidebar-plus is the single binary behind the plugin's startup
// hook, event hooks, jump action, and jump popup.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/amberpixels/herdr-sidebar-plus/internal/herdr"
	"github.com/amberpixels/herdr-sidebar-plus/internal/labels"
	"github.com/amberpixels/herdr-sidebar-plus/internal/sidebar"
)

const (
	pluginID      = "amberpixels.sidebar-plus"
	jumpPaneID    = "jump"
	rulesFileName = "rules.toml"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: herdr-sidebar-plus refresh|open-jump|jump")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "refresh":
		err = refresh()
	case "open-jump":
		err = withClient(func(c *herdr.Client) error { return c.OpenPluginPane(pluginID, jumpPaneID) })
	case "jump":
		err = jump()
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "herdr-sidebar-plus:", err)
		os.Exit(1)
	}
}

func withClient(fn func(*herdr.Client) error) error {
	client, err := herdr.Dial()
	if err != nil {
		return err
	}
	return fn(client)
}

// refresh runs from the startup hook and every event hook. Hooks fire in bursts
// and run concurrently, so an exclusive lock serializes them: each run reads the
// session only after the previous one wrote, and the last one leaves it current.
func refresh() error {
	rules, err := labels.Load(filepath.Join(os.Getenv("HERDR_PLUGIN_CONFIG_DIR"), rulesFileName))
	if err != nil {
		return err
	}
	unlock, err := lock(filepath.Join(os.Getenv("HERDR_PLUGIN_STATE_DIR"), "refresh.lock"))
	if err != nil {
		return err
	}
	defer unlock()
	return withClient(func(c *herdr.Client) error { return sidebar.Refresh(c, rules) })
}

func lock(path string) (func(), error) {
	// The path is under HERDR_PLUGIN_STATE_DIR, which Herdr sets for this plugin.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600) //nolint:gosec // trusted plugin state dir
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		return nil, err
	}
	return func() { _ = file.Close() }, nil
}
