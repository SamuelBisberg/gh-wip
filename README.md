<div align="center">

# gh-wip

*Context-switch without losing your train*

[![CI](https://github.com/SamuelBisberg/gh-wip/actions/workflows/ci.yml/badge.svg)](https://github.com/SamuelBisberg/gh-wip/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/SamuelBisberg/gh-wip?sort=semver)](https://github.com/SamuelBisberg/gh-wip/releases)

[Install](#install) • [Usage](#usage) • [Configuration](#configuration) • [How it works](#how-it-works)

</div>

A [GitHub CLI](https://cli.github.com) extension that snapshots your uncommitted changes onto a `wip/<timestamp>` branch on `origin` and hands you back a clean tree - pull it down again later, on this machine or any other that can reach the same remote. A single static binary: no `fzf`, no Node, no Python, just your existing `gh auth login` session and `git`.

![gh wip push, then gh wip pull, restoring the same changes](assets/demo.gif)

## Install

```sh
gh extension install SamuelBisberg/gh-wip
```

Upgrade later with `gh extension upgrade wip`.

## Usage

### `gh wip push`

Stashes your uncommitted changes onto a new remote branch and restores your working branch to a clean state.

```console
$ git status --short
 M search.go
$ gh wip push
✓ Captured WIP as origin/wip/20260831-140501
ℹ Restore it later with: gh wip pull
$ git status --short
$
```

> [!TIP]
> Configure an [AI driver](#ai-commit-summaries) and the branch's commit message becomes a real summary of the diff instead of a timestamp.

### `gh wip pull`

Fetches every `wip/*` branch from `origin`, lets you filter and pick one, and squash-merges it into your working directory as staged, uncommitted changes - mirroring exactly the state `push` captured.

```sh
gh wip pull
```

`pull` refuses to run against a dirty working tree, so it never clobbers changes you haven't captured yet. Pass `--delete` to skip the confirmation and always delete the remote branch once it's merged cleanly. If the merge conflicts, gh-wip leaves the conflict markers for you to resolve rather than guessing - nothing is auto-aborted.

### `gh wip cleanup`

Checks off any number of `wip/*` branches from a list and deletes them all - on `origin`, and locally too if a matching branch still exists there.

```sh
gh wip cleanup
```

Space toggles a branch, enter submits the selection. Since this is permanent, gh-wip won't act on it until you type `yes` at the confirmation prompt - anything else, including just pressing enter, cancels.

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
| `ai.customCommand` | shell command                         | *(unset)* | Used when `ai.driver` is `custom`.                                              |
| `pull.autoDelete`  | `true`, `false`                       | `false`   | Skip the confirmation and always delete the remote branch after a clean `pull`. |
| `ui.color`         | `auto`, `always`, `never`             | `auto`    | Override terminal color detection.                                              |

### AI commit summaries

gh-wip never talks to an AI API directly - it shells out to a CLI you already have installed, so there's nothing to authenticate and no keys for gh-wip to manage:

- **`claude`** runs `claude -p "<prompt>"` ([Claude Code](https://docs.claude.com/en/docs/claude-code)'s headless mode).
- **`copilot`** runs `copilot -p "<prompt>" -s` ([GitHub Copilot CLI](https://docs.github.com/copilot/how-tos/copilot-cli)'s headless mode).
- **`custom`** runs whatever shell command you set as `ai.customCommand`, piping the diff to its stdin and reading the summary back from stdout - the escape hatch for any AI tool gh-wip doesn't wire up directly:

  ```sh
  gh wip config set ai.driver custom
  gh wip config set ai.customCommand 'my-openai-commit-summarizer'
  ```

> [!NOTE]
> If the chosen driver's binary isn't installed, or the call errors out or times out (45s), `push` prints a warning and falls back to a plain `WIP: snapshot of <branch> at <time>` message - a summarizer failure never blocks a push.

## How it works

- **Git operations** shell out to your system `git`, using whatever credential helper `gh auth login` already configured - gh-wip never sees or stores a token itself.
- **`gh wip push`** creates `wip/<UTC timestamp>`, stages everything, commits, pushes to `origin`, and checks your original branch back out. If the push fails (e.g. a network blip), the commit stays safe locally on the WIP branch and gh-wip prints the exact `git push` to retry.
- **`gh wip pull`** squash-merges your selection - no merge commit, changes land staged and uncommitted, same shape they were in before `push`.
- The interactive picker and prompts are [charmbracelet/huh](https://github.com/charmbracelet/huh); styling is [lipgloss](https://github.com/charmbracelet/lipgloss). Everything renders in-process - no subprocess.

Releases are built by [`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile) whenever a version tag is pushed, producing binaries for Linux, macOS, and Windows - that's what `gh extension install`/`upgrade` fetch.

Want to contribute? See [CONTRIBUTING.md](CONTRIBUTING.md).
