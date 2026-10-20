// Package sidebar computes the custom sidebar tokens this plugin reports:
// $label on every workspace and agent pane, and $num on every agent pane.
package sidebar

import (
	"strconv"

	"github.com/amberpixels/herdr-sidebar-plus/internal/herdr"
	"github.com/amberpixels/herdr-sidebar-plus/internal/labels"
)

const (
	// Source identifies this plugin's metadata reports.
	Source = "amberpixels.sidebar-plus"

	LabelToken = "label"
	NumToken   = "num"
)

// Patch is one metadata report: a workspace patch when PaneID is empty,
// otherwise a pane patch.
type Patch struct {
	WorkspaceID string
	PaneID      string
	Tokens      map[string]*string
}

// Plan returns only the patches whose values differ from what the server
// already holds, so an unchanged session costs no writes or redraws.
func Plan(workspaces []herdr.Workspace, agents []herdr.Agent, rules *labels.Rules) []Patch {
	var patches []Patch
	workspaceLabels := make(map[string]string, len(workspaces))
	for _, ws := range workspaces {
		var repoName, repoRoot string
		if ws.Worktree != nil {
			repoName, repoRoot = ws.Worktree.RepoName, ws.Worktree.RepoRoot
		}
		label := rules.Label(ws.Label, repoName, repoRoot)
		workspaceLabels[ws.ID] = label
		if tokens := changed(ws.Tokens, map[string]string{LabelToken: label}); tokens != nil {
			patches = append(patches, Patch{WorkspaceID: ws.ID, Tokens: tokens})
		}
	}
	// Agent order matches Herdr's agent panel in "spaces" sort, which is also the
	// order the built-in focus_agent shortcuts use.
	for i, agent := range agents {
		want := map[string]string{NumToken: strconv.Itoa(i + 1)}
		if label, ok := workspaceLabels[agent.WorkspaceID]; ok {
			want[LabelToken] = label
		}
		if tokens := changed(agent.Tokens, want); tokens != nil {
			patches = append(patches, Patch{WorkspaceID: agent.WorkspaceID, PaneID: agent.PaneID, Tokens: tokens})
		}
	}
	return patches
}

func changed(current, want map[string]string) map[string]*string {
	var tokens map[string]*string
	for key, value := range want {
		if current[key] == value {
			continue
		}
		if tokens == nil {
			tokens = make(map[string]*string, len(want))
		}
		tokens[key] = &value
	}
	return tokens
}

// Refresh reads the session and reports every changed token.
func Refresh(client *herdr.Client, rules *labels.Rules) error {
	workspaces, err := client.Workspaces()
	if err != nil {
		return err
	}
	agents, err := client.Agents()
	if err != nil {
		return err
	}
	for _, patch := range Plan(workspaces, agents, rules) {
		if patch.PaneID == "" {
			err = client.ReportWorkspaceTokens(patch.WorkspaceID, Source, patch.Tokens)
		} else {
			err = client.ReportPaneTokens(patch.PaneID, Source, patch.Tokens)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
