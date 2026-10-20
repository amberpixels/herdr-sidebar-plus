package labels

import (
	"os"
	"path/filepath"
	"testing"
)

const acme = `
[[project]]
repo = "acme"
match = '^(acme\.)?PROJ-\d+[/-]'
steps = [
  { find = '^acme\.', replace = '' },
  { find = '^PROJ-(\d+)[/-]', replace = '${1}: ' },
  { find = '-', replace = ' ' },
]
`

func load(t *testing.T, body string) *Rules {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rules.toml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	rules, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func TestLabel(t *testing.T) {
	rules := load(t, acme)
	cases := []struct {
		name, repo, want string
	}{
		{"PROJ-1234/fix-invoice-rounding-on-refunds", "acme", "1234: fix invoice rounding on refunds"},
		{"acme.PROJ-1250-add-dark-mode-toggle", "acme", "1250: add dark mode toggle"},
		// The gate keeps names without a task prefix untouched.
		{"acme", "acme", "acme"},
		{"docs-site", "", "docs-site"},
		// Rules are scoped to their repo.
		{"PROJ-1/other-repo", "billing", "PROJ-1/other-repo"},
	}
	for _, tc := range cases {
		if got := rules.Label(tc.name, tc.repo, ""); got != tc.want {
			t.Errorf("Label(%q, %q) = %q, want %q", tc.name, tc.repo, got, tc.want)
		}
	}
}

func TestLabelPathScopeAndBlankFallback(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(home, "code", "app")
	rules := load(t, `
[[project]]
path = "~/code/app"
steps = [{ find = '.*', replace = '' }]
`)
	if got := rules.Label("feature-x", "app", root); got != "feature-x" {
		t.Errorf("blank rewrite should fall back to the name, got %q", got)
	}
	rules = load(t, `
[[project]]
path = "~/code/app"
steps = [{ find = '^feature-', replace = '' }]
`)
	if got := rules.Label("feature-x", "app", root+string(filepath.Separator)); got != "x" {
		t.Errorf("path scope with ~ and trailing slash, got %q", got)
	}
	if got := rules.Label("feature-x", "app", "/elsewhere/app"); got != "feature-x" {
		t.Errorf("other path must not match, got %q", got)
	}
}

func TestLoadMissingFileAndBadRegex(t *testing.T) {
	rules, err := Load(filepath.Join(t.TempDir(), "absent.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := rules.Label("a-b", "", ""); got != "a-b" {
		t.Errorf("no rules should keep the name, got %q", got)
	}
	path := filepath.Join(t.TempDir(), "rules.toml")
	if err := os.WriteFile(path, []byte(`[[project]]
match = '('`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("invalid regex should fail to load")
	}
}
