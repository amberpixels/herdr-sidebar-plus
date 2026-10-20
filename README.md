# herdr-sidebar-plus

A [Herdr](https://herdr.dev) plugin that makes the sidebar easier to scan and drive from the keyboard:

- **Numbered agents.** Every agent gets a `$num` token: its position in the agent panel, past 9 too.
- **Two-digit jump.** `prefix+0` opens a tiny prompt; type two digits to focus that agent (`0 7`, `1 2`).
- **Readable labels.** A `$label` token on every space and agent, rewritten by per-project regex rules.
  `PROJ-1234/fix-invoice-rounding-on-refunds` becomes `1234: fix invoice rounding on refunds`.

```text
 ○ acme · main              ○ 1 · acme
   ├─ ○ 1234: fix inv…      ○ 2 · docs-site
   └─ ○ 1250: add dar…      ○ 3 · 1234: fix invoic…
```

Requires Herdr 0.9.5+ on Linux or macOS, and Go to build.

## Install

```bash
herdr plugin install amberpixels/herdr-sidebar-plus
```

Or from a checkout:

```bash
just link   # builds bin/herdr-sidebar-plus, then `herdr plugin link .`
```

## Configure Herdr

The plugin only reports tokens. Show them by using `$label` and `$num` in your sidebar rows, and bind the jump:

```toml
[ui]
agent_panel_sort = "spaces"   # numbers follow this order

[ui.sidebar.spaces]
rows = [["state_icon", "$label", "branch", "git_status"]]

[ui.sidebar.agents]
rows = [["state_icon", "$num", "$label"]]

[[keys.command]]
key = "prefix+0"
type = "plugin_action"
command = "amberpixels.sidebar-plus.jump"
description = "jump to agent by number"
```

In the jump prompt, any key other than a digit (Esc included) closes it without moving focus.

## Label rules

Rules live in `rules.toml` in the plugin's config directory:

```bash
cp rules.example.toml "$(herdr plugin config-dir amberpixels.sidebar-plus)/rules.toml"
```

```toml
[[project]]
repo = "acme"                        # worktree repo name; or `path = "~/code/acme"`
match = '^(acme\.)?PROJ-\d+[/-]'    # only rewrite names that look like tasks
steps = [
  { find = '^acme\.', replace = '' },
  { find = '^PROJ-(\d+)[/-]', replace = '${1}: ' },
  { find = '-', replace = ' ' },
]
```

- `repo` matches the repo name and `path` the repo root of a worktree-backed workspace. With neither, the project applies to every workspace.
- `match` gates the project; without it, the steps always run.
- Steps are Go regex replacements applied in order. Write capture groups as `${1}`.
- The first project that applies wins. Names no project rewrites keep their original text, so no row goes blank.

Rules are read on every refresh. After editing them, run `just refresh` (or `herdr plugin action invoke amberpixels.sidebar-plus.refresh`).

## How it works

A startup hook and hooks on workspace, tab, pane, and worktree lifecycle events run `herdr-sidebar-plus refresh`. It reads `workspace.list` and `agent.list`, computes labels and numbers, and reports only the values that changed. A lock file serializes refreshes that fire in a burst.

## Limits

- Herdr draws ` · ` between tokens, so a numbered agent reads `○ 1 · name`.
- Herdr trims token values, so two-digit numbers cannot be padded into a right-aligned column.
- Numbers follow `agent.list` order, which matches the agent panel in `spaces` sort. In `priority` sort the panel reorders on every status change, and the numbers will not follow it.
- Labels update when an event fires, so a rename shows the old text for a moment.

## License

[MIT](LICENSE)
