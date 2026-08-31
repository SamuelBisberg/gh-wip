# Contributing to gh-wip

Thanks for taking the time to contribute!

## Prerequisites

- [Go](https://go.dev) 1.25+ (see `go.mod`)
- [git](https://git-scm.com)
- [GitHub CLI](https://cli.github.com) (`gh`), to run the extension locally

## Getting started

```sh
git clone https://github.com/SamuelBisberg/gh-wip.git
cd gh-wip
go build ./...
```

Install it as a local `gh` extension to try your changes end-to-end:

```sh
gh extension install .
gh wip push
```

> [!TIP]
> After changing code, `gh extension remove wip && gh extension install .` picks up a fresh build.

## Making changes

```sh
go build ./...
go vet ./...
gofmt -l .          # should print nothing
go test -race ./...
```

All four must pass before opening a PR - CI runs the same checks.

### Project layout

| Package      | Responsibility                                                              |
| ------------ | --------------------------------------------------------------------------- |
| `cmd/`       | The `gh wip` command tree ([cobra](https://github.com/spf13/cobra))         |
| `pkg/git`    | Thin wrapper around the system `git` binary                                 |
| `pkg/ai`     | Pluggable commit-message summarizers (Claude, Copilot, or a custom command) |
| `pkg/config` | Persisted settings (`~/.config/gh/wip.json`)                                |
| `pkg/tui`    | Terminal styling, spinners, and interactive prompts                         |

### Tests

`pkg/git`'s tests create real, disposable git repositories (and bare "remotes") in temp directories - no network access, no mocking of `git` itself. Follow that pattern for new git-backed behavior rather than mocking the package.

> [!WARNING]
> Never run `gh wip push`/`gh wip pull` from a checkout of this repo itself - they create and push real branches to whatever `origin` is configured. Test against a throwaway repo instead.

## Submitting a pull request

1. Fork the repo and create a branch off `main`.
2. Make your change, with tests.
3. Run the checks above.
4. Open a PR - the template will guide you through the rest.

## Reporting bugs / requesting features

Please use the issue templates; they ask for just enough context to act on a report quickly.
