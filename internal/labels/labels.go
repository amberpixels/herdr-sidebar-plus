// Package labels rewrites workspace names into shorter sidebar labels using
// per-project regex rules.
package labels

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the user-editable rules file.
type Config struct {
	Projects []Project `toml:"project"`
}

// Project scopes rules to one repository. Repo matches the workspace's repo name
// and Path its repo root; with neither set the project applies everywhere.
// Steps run in order only when Match (if set) matches the original name. Icon
// marks every workspace of the repo, whatever Match says.
type Project struct {
	Repo  string `toml:"repo"`
	Path  string `toml:"path"`
	Match string `toml:"match"`
	Icon  string `toml:"icon"`
	Steps []Step `toml:"steps"`
}

// Step is one regex replacement; Replace uses Go's ${1} expansion syntax.
type Step struct {
	Find    string `toml:"find"`
	Replace string `toml:"replace"`
}

// Rules is a compiled Config.
type Rules struct {
	projects []project
}

type project struct {
	repo  string
	path  string
	match *regexp.Regexp
	icon  string
	steps []step
}

type step struct {
	find    *regexp.Regexp
	replace string
}

// Load reads and compiles the rules file. A missing file yields no rules, so
// every label falls back to the plain workspace name.
func Load(path string) (*Rules, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &Rules{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	rules, err := Compile(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return rules, nil
}

func Compile(cfg Config) (*Rules, error) {
	rules := &Rules{}
	for i, p := range cfg.Projects {
		compiled := project{repo: p.Repo, path: expandHome(p.Path), icon: strings.TrimSpace(p.Icon)}
		if p.Match != "" {
			re, err := regexp.Compile(p.Match)
			if err != nil {
				return nil, fmt.Errorf("project %d match: %w", i+1, err)
			}
			compiled.match = re
		}
		for j, s := range p.Steps {
			re, err := regexp.Compile(s.Find)
			if err != nil {
				return nil, fmt.Errorf("project %d step %d: %w", i+1, j+1, err)
			}
			compiled.steps = append(compiled.steps, step{find: re, replace: s.Replace})
		}
		rules.projects = append(rules.projects, compiled)
	}
	return rules, nil
}

// Label returns the display label for a workspace name. repoName and repoRoot
// are empty for workspaces outside a Git worktree. The first project that
// applies wins; an empty result falls back to the name so no row goes blank.
func (r *Rules) Label(name, repoName, repoRoot string) string {
	for _, p := range r.projects {
		if !p.scopes(repoName, repoRoot) {
			continue
		}
		if p.match != nil && !p.match.MatchString(name) {
			continue
		}
		label := name
		for _, s := range p.steps {
			label = s.find.ReplaceAllString(label, s.replace)
		}
		if label = strings.TrimSpace(label); label != "" {
			return label
		}
		return name
	}
	return name
}

// Icon returns the icon of the first project scoped to the repo that sets one,
// or "" when none does.
func (r *Rules) Icon(repoName, repoRoot string) string {
	for _, p := range r.projects {
		if p.icon != "" && p.scopes(repoName, repoRoot) {
			return p.icon
		}
	}
	return ""
}

func (p project) scopes(repoName, repoRoot string) bool {
	if p.repo != "" && p.repo != repoName {
		return false
	}
	return p.path == "" || filepath.Clean(p.path) == filepath.Clean(repoRoot)
}

func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}
