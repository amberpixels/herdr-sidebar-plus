package main

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/term"

	"github.com/amberpixels/herdr-sidebar-plus/internal/herdr"
	"github.com/amberpixels/herdr-sidebar-plus/internal/sidebar"
)

// jump runs inside the popup: it reads exactly two digits and focuses the agent
// whose $num token matches. Any other key closes the popup without focusing.
func jump() error {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(fd, state) }()

	prompt := func(typed string) { fmt.Printf("\r\x1b[2K agent %s", typed) }
	prompt("__")
	number := 0
	for i := range 2 {
		digit, ok := readDigit()
		if !ok {
			return nil
		}
		number = number*10 + digit
		if i == 0 {
			prompt(strconv.Itoa(digit) + "_")
		}
	}
	if number == 0 {
		return nil
	}
	return withClient(func(c *herdr.Client) error {
		agents, err := c.Agents()
		if err != nil {
			return err
		}
		if pane := agentPane(agents, number); pane != "" {
			return c.FocusAgent(pane)
		}
		return nil
	})
}

func readDigit() (int, bool) {
	var buf [1]byte
	if n, err := os.Stdin.Read(buf[:]); err != nil || n == 0 || buf[0] < '0' || buf[0] > '9' {
		return 0, false
	}
	return int(buf[0] - '0'), true
}

// agentPane prefers the number the sidebar shows, falling back to list order
// when a refresh has not reported numbers yet.
func agentPane(agents []herdr.Agent, number int) string {
	want := strconv.Itoa(number)
	for _, agent := range agents {
		if agent.Tokens[sidebar.NumToken] == want {
			return agent.PaneID
		}
	}
	if number <= len(agents) {
		return agents[number-1].PaneID
	}
	return ""
}
