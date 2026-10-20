package sidebar

import (
	"testing"

	"github.com/amberpixels/herdr-sidebar-plus/internal/herdr"
	"github.com/amberpixels/herdr-sidebar-plus/internal/labels"
)

func TestPlan(t *testing.T) {
	rules, err := labels.Compile(labels.Config{Projects: []labels.Project{{
		Repo:  "orchard",
		Steps: []labels.Step{{Find: `^ORC-(\d+)/`, Replace: "${1}: "}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	workspaces := []herdr.Workspace{
		{ID: "w1", Label: "orchard", Worktree: &herdr.Worktree{RepoName: "orchard"}},
		{
			ID: "w2", Label: "ORC-1/fix", Worktree: &herdr.Worktree{RepoName: "orchard"},
			Tokens: map[string]string{LabelToken: "1: fix"},
		},
	}
	agents := []herdr.Agent{
		{PaneID: "w2:p1", WorkspaceID: "w2", Tokens: map[string]string{NumToken: "1", LabelToken: "1: fix"}},
		{PaneID: "w1:p1", WorkspaceID: "w1", Tokens: map[string]string{NumToken: "1"}},
	}

	patches := Plan(workspaces, agents, rules)

	// w2 and its agent already hold the right values; only w1 and the renumbered
	// second agent need writes.
	if len(patches) != 2 {
		t.Fatalf("patches = %+v", patches)
	}
	if p := patches[0]; p.PaneID != "" || p.WorkspaceID != "w1" || *p.Tokens[LabelToken] != "orchard" {
		t.Errorf("workspace patch = %+v", p)
	}
	p := patches[1]
	if p.PaneID != "w1:p1" || *p.Tokens[NumToken] != "2" || *p.Tokens[LabelToken] != "orchard" {
		t.Errorf("agent patch = %+v", p)
	}
}
