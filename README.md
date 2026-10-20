# herdr-sidebar-plus

A [Herdr](https://herdr.dev) plugin that makes the sidebar easier to scan:

- **Readable labels.** A `$label` token on every space and agent, rewritten by per-project regex rules.
  `ORC-1234/fix-invoice-rounding-on-refunds` becomes `1234: fix invoice rounding on refunds`.
- **Project icons.** An `$icon` token on every space and agent of a project, so a repo and all its worktrees share one glyph.

```text
 ○ orchard · main           ○ orchard
   ├─ ○ 1234: fix inv…      ○ orchard-docs
   └─ ○ 1250: add dar…      ○ 1234: fix invoic…
```

Requires Herdr 0.9.5+ on Linux or macOS, and Go to build. Drawing icons in place of the state dot needs a Herdr build with the `projects` indicator style, which is not in stock Herdr yet.

## Install

```bash
herdr plugin install amberpixels/herdr-sidebar-plus
```

Or from a checkout:

```bash
just link   # builds bin/herdr-sidebar-plus, then `herdr plugin link .`
```

## Configure Herdr

The plugin only reports tokens. Show `$label` in your sidebar rows:

```toml
[ui.sidebar.spaces]
rows = [["state_icon", "$label", "branch", "git_status"]]

[ui.sidebar.agents]
rows = [["state_icon", "$label"]]
```

To draw each row's `$icon` as its state indicator, colored by agent state, use the `projects` style. Rows of projects without an icon keep the usual dots:

```toml
[ui]
status_indicators = "projects"
```

On stock Herdr, place `$icon` in a row like any other token instead.

## Follow `cd`

Herdr renames a space after its shell changes directory. Stock Herdr fires no event a plugin can hook for that, so without help `$label` and `$icon` catch up only on the next focus change. For zsh, add this hook to `.zshrc` to refresh on every `cd` inside Herdr:

```zsh
if [[ -n $HERDR_ENV ]]; then
  autoload -Uz add-zsh-hook
  _herdr_sidebar_plus_chpwd() {
    "${HERDR_BIN_PATH:-herdr}" plugin action invoke amberpixels.sidebar-plus.refresh >/dev/null 2>&1 &!
  }
  add-zsh-hook chpwd _herdr_sidebar_plus_chpwd
fi
```

Herdr builds with the `projects` style fire `workspace.updated` when a `cd` renames a space, so they need no hook.

## Rules

Rules live in `rules.toml` in the plugin's config directory:

```bash
cp rules.example.toml "$(herdr plugin config-dir amberpixels.sidebar-plus)/rules.toml"
```

```toml
[[project]]
repo = "orchard"                    # repo name; or `path = "~/code/orchard"`
icon = "󰔱"                          # any single glyph, Nerd Font icons included
match = '^(orchard\.)?ORC-\d+[/-]'  # only rewrite names that look like tasks
steps = [
  { find = '^orchard\.', replace = '' },
  { find = '^ORC-(\d+)[/-]', replace = '${1}: ' },
  { find = '-', replace = ' ' },
]
```

- `repo` matches the repo name and `path` the main repo root. A linked worktree reports its main repo, so both cover every worktree. With neither, the project applies to every workspace.
- `icon` marks every space and agent of the project, whatever `match` says. The first project with an icon wins.
- `match` gates the label steps; without it, the steps always run.
- Steps are Go regex replacements applied in order. Write capture groups as `${1}`.
- The first project that applies wins. Names no project rewrites keep their original text, so no row goes blank.

Rules are read on every refresh. After editing them, run `just refresh` (or `herdr plugin action invoke amberpixels.sidebar-plus.refresh`).

## How it works

A startup hook and hooks on workspace, tab, pane, and worktree lifecycle and focus events run `herdr-sidebar-plus refresh`. It reads `workspace.list` and `agent.list`, computes labels and icons, and reports only the values that changed. A lock file serializes refreshes that fire in a burst.

Stock Herdr reports a workspace's repo only when it was opened as a worktree, so a plain space inside a repo gets no repo-scoped label or icon there. Herdr builds with the `projects` style report the repo of every workspace.

## Limits

- Labels update when an event fires, so a rename shows the old text for a moment.
- On stock Herdr, a space renamed by a `cd` keeps its old label until the next focus change, unless the shell runs the [`cd` hook](#follow-cd).

## License

[MIT](LICENSE)
