package sidebar

import (
	"testing"

	"github.com/amberpixels/herdr-sidebar-plus/internal/herdr"
	"github.com/amberpixels/herdr-sidebar-plus/internal/labels"
)

func TestPlan(t *testing.T) {
	rules, err := labels.Compile(labels.Config{Projects: []labels.Project{{
		Repo:  "orchard",
		Icon:  "@",
		Steps: []labels.Step{{Find: `^ORC-(\d+)/`, Replace: "${1}: "}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	orchard := &herdr.Repo{RepoName: "orchard"}
	workspaces := []herdr.Workspace{
		// Only the live repo knows this plain space is in orchard.
		{ID: "w1", Label: "orchard", Repo: orchard},
		{
			ID: "w2", Label: "ORC-1/fix", Worktree: orchard,
			Tokens: map[string]string{LabelToken: "1: fix", IconToken: "@"},
		},
		// A space that left the repo drops its stale icon.
		{ID: "w3", Label: "notes", Tokens: map[string]string{LabelToken: "notes", IconToken: "@"}},
	}
	agents := []herdr.Agent{
		{PaneID: "w2:p1", WorkspaceID: "w2", Tokens: map[string]string{LabelToken: "1: fix", IconToken: "@"}},
		{PaneID: "w1:p1", WorkspaceID: "w1"},
	}

	patches := Plan(workspaces, agents, rules)

	// w2 and its agent already hold the right values.
	if len(patches) != 3 {
		t.Fatalf("patches = %+v", patches)
	}
	if p := patches[0]; p.WorkspaceID != "w1" || *p.Tokens[LabelToken] != "orchard" || *p.Tokens[IconToken] != "@" {
		t.Errorf("w1 patch = %+v", p)
	}
	if p := patches[1]; p.WorkspaceID != "w3" || len(p.Tokens) != 1 || p.Tokens[IconToken] != nil {
		t.Errorf("w3 patch should only clear the icon, got %+v", p)
	}
	if p := patches[2]; p.PaneID != "w1:p1" || *p.Tokens[LabelToken] != "orchard" || *p.Tokens[IconToken] != "@" {
		t.Errorf("agent patch = %+v", p)
	}
}
