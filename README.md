# gh-wip

[![CI](https://github.com/SamuelBisberg/gh-wip/actions/workflows/ci.yml/badge.svg)](https://github.com/SamuelBisberg/gh-wip/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/SamuelBisberg/gh-wip?sort=semver)](https://github.com/SamuelBisberg/gh-wip/releases)

A [GitHub CLI](https://cli.github.com) extension for context-switching without losing your place. `gh wip push` snapshots whatever's uncommitted in your working directory onto a `wip/<timestamp>` branch on `origin` and hands you back a clean tree; `gh wip pull` lets you pick one of those branches from an interactive list and merges it back in — on this machine, another machine, or a teammate's, as long as they can reach the same remote.

It's a single static binary. No fzf, no Node, no Python — it inherits your existing `gh auth login` session and talks to `git` directly.

## Install

```sh
gh extension install SamuelBisberg/gh-wip
```

Upgrade later with `gh extension upgrade wip`.

## Usage

### `gh wip push`

Stashes your uncommitted changes onto a new remote branch and restores your working branch to a clean state.

```sh
$ gh wip push
Capturing changes on wip/20260831-140501...
Pushing wip/20260831-140501 to origin...
✓ Captured WIP as origin/wip/20260831-140501
ℹ Restore it later with: gh wip pull
```

If an AI driver is configured (see [Configuration](#configuration)), the commit message is a generated summary of the diff instead of a timestamp. If generation fails or isn't configured, it falls back to a plain `WIP: snapshot of <branch> at <time>` message — push never blocks on AI.

### `gh wip pull`

Fetches every `wip/*` branch from `origin`, lets you fuzzy-filter and pick one, and squash-merges it into your current working directory as staged, uncommitted changes — mirroring exactly the state `push` captured.

```sh
$ gh wip pull
Fetching from origin...
? Select a WIP branch to pull
> wip/20260831-140501  —  3 minutes ago  —  Add pagination to the search results (Jane Doe)
  wip/20260830-091200  —  yesterday      —  WIP: snapshot of main at 2026-08-30 09:12:00 UTC (Jane Doe)
✓ Merged wip/20260831-140501 — changes are staged in your working directory, uncommitted.
? Delete remote branch wip/20260831-140501 now that it's merged? (Y/n)
```

`gh wip pull` refuses to run against a dirty working tree, so it never clobbers changes you haven't captured yet. Pass `--delete` to skip the confirmation and always delete the remote branch once it's merged cleanly; set `pull.autoDelete` (below) to make that the default. If the merge conflicts, gh-wip leaves the conflict markers for you to resolve rather than guessing — nothing is auto-aborted.

### `gh wip config`

```sh
gh wip config list                     # show every setting and where it's stored
gh wip config get ai.driver
gh wip config set ai.driver claude
```

## Configuration

Settings live in `~/.config/gh/wip.json` (or your platform's config dir equivalent) and are managed with `gh wip config`.

| Key                | Values                                | Default   |                                                                                 |
| ------------------ | ------------------------------------- | --------- | ------------------------------------------------------------------------------- |
| `ai.driver`        | `none`, `claude`, `copilot`, `custom` | `none`    | Which AI summarizer to use for `push` commit messages.                          |
| `ai.customCommand` | shell command                         | *(unset)* | Used when `ai.driver` is `custom` — see below.                                  |
| `pull.autoDelete`  | `true`, `false`                       | `false`   | Skip the confirmation and always delete the remote branch after a clean `pull`. |
| `ui.color`         | `auto`, `always`, `never`             | `auto`    | Override terminal color detection.                                              |

### AI-generated commit summaries

gh-wip never talks to an AI API directly — it shells out to a CLI you already have installed, so there's nothing to authenticate and no keys for gh-wip to manage:

- **`claude`** runs [`claude -p "<prompt>"`](https://docs.claude.com/en/docs/claude-code) (Claude Code's headless mode). Requires the `claude` CLI on your `PATH`.
- **`copilot`** runs `copilot -p "<prompt>" -s` ([GitHub Copilot CLI](https://docs.github.com/copilot/how-tos/copilot-cli)'s headless mode). Requires the `copilot` CLI on your `PATH`.
- **`custom`** runs whatever shell command you set as `ai.customCommand`, piping the diff to its stdin and reading the summary back from stdout. Point it at your own script to wire up OpenAI, a local model, or anything else:

  ```sh
  gh wip config set ai.driver custom
  gh wip config set ai.customCommand 'my-openai-commit-summarizer'
  ```

If the chosen driver's binary isn't installed, or the call errors out or times out (45s), `push` prints a warning and falls back to the default timestamp message — it never fails the push.

## How it works

- **Git operations** shell out to your system `git`, so they use whatever credential helper `gh auth login` already configured — gh-wip never sees or stores a token.
- **`gh wip push`**: checks for uncommitted changes, generates a message (AI or timestamp), creates `wip/<UTC timestamp>`, stages everything, commits, pushes to `origin`, and checks your original branch back out. If the push fails (e.g. a network blip), the commit stays safe locally on the WIP branch and gh-wip tells you the exact `git push` to retry.
- **`gh wip pull`**: fetches `origin`, lists remote branches under `wip/`, and squash-merges your selection — no merge commit, changes land staged and uncommitted, same shape they were in before `push`.
- The interactive picker and prompts are [charmbracelet/huh](https://github.com/charmbracelet/huh); styling is [lipgloss](https://github.com/charmbracelet/lipgloss). Everything renders in-process — no `fzf` subprocess.

## Development

```sh
go build ./...
go vet ./...
gofmt -l .
```

Releases are built by [`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile) whenever a `v*.*.*` tag is pushed, producing binaries for Linux, macOS, and Windows (amd64/arm64) and attaching them to the GitHub release — that's what `gh extension install`/`upgrade` fetch.
