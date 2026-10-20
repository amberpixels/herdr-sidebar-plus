// Package sidebar computes the custom sidebar tokens this plugin reports:
// $label and $icon on every workspace and agent pane.
package sidebar

import (
	"github.com/amberpixels/herdr-sidebar-plus/internal/herdr"
	"github.com/amberpixels/herdr-sidebar-plus/internal/labels"
)

const (
	// Source identifies this plugin's metadata reports.
	Source = "amberpixels.sidebar-plus"

	LabelToken = "label"
	// IconToken is the token Herdr's "projects" indicator style draws.
	IconToken = "icon"
)

// Patch is one metadata report: a workspace patch when PaneID is empty,
// otherwise a pane patch. A nil token value clears that token.
type Patch struct {
	WorkspaceID string
	PaneID      string
	Tokens      map[string]*string
}

// Plan returns only the patches whose values differ from what the server
// already holds, so an unchanged session costs no writes or redraws.
func Plan(workspaces []herdr.Workspace, agents []herdr.Agent, rules *labels.Rules) []Patch {
	var patches []Patch
	workspaceTokens := make(map[string]map[string]string, len(workspaces))
	for _, ws := range workspaces {
		repo := ws.CurrentRepo()
		want := map[string]string{
			LabelToken: rules.Label(ws.Label, repo.RepoName, repo.RepoRoot),
			IconToken:  rules.Icon(repo.RepoName, repo.RepoRoot),
		}
		workspaceTokens[ws.ID] = want
		if tokens := changed(ws.Tokens, want); tokens != nil {
			patches = append(patches, Patch{WorkspaceID: ws.ID, Tokens: tokens})
		}
	}
	for _, agent := range agents {
		want, ok := workspaceTokens[agent.WorkspaceID]
		if !ok {
			continue
		}
		if tokens := changed(agent.Tokens, want); tokens != nil {
			patches = append(patches, Patch{WorkspaceID: agent.WorkspaceID, PaneID: agent.PaneID, Tokens: tokens})
		}
	}
	return patches
}

// changed returns the tokens to write: values that differ, and a nil for every
// wanted-empty token the server still holds.
func changed(current, want map[string]string) map[string]*string {
	var tokens map[string]*string
	for key, value := range want {
		held, ok := current[key]
		if held == value || (value == "" && !ok) {
			continue
		}
		if tokens == nil {
			tokens = make(map[string]*string, len(want))
		}
		if value == "" {
			tokens[key] = nil
		} else {
			tokens[key] = &value
		}
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
