# Contributing to Pressie

Pressie is an open-source, agent-friendly gift tracking CLI. Contributions are welcome — bug fixes, new commands, plugin improvements, web UI enhancements, documentation.

## AI-Generated Contributions

AI-generated code is welcome. If you use an AI tool (Claude, Copilot, OpenClaw, etc.) to write code, that's fine. **You are responsible for the result:**

- Run the tests yourself and verify they pass.
- Test the actual behavior — don't just trust the AI's claim that it works.
- Review the diff before submitting. Understand what changed and why.
- If the AI introduced a pattern or abstraction, make sure it fits the existing codebase.

The human submitting the PR is the author, regardless of which tools were used to write the code.

## Getting Started

1. **Fork** the repository.
2. **Clone** your fork locally.
3. **Create a branch** for your change: `git checkout -b my-feature`.
4. **Build** to verify the project compiles: `go build ./...`
5. **Run tests**: `go test ./...`
6. **Make your changes.** Follow the patterns in the existing code.
7. **Test your changes**: `go test ./...` and `go vet ./...` must pass.
8. **Commit** with a clear message describing what and why.
9. **Push** to your fork and open a PR against `main`.

## Testing Requirements

All PRs must meet these standards:

- **`go test ./...` passes** — no exceptions.
- **`go vet ./...` passes** — no exceptions.
- **New features need tests.** If you add a command, a handler, or a helper function, write tests that exercise the observable behavior. Don't test plumbing — test outcomes.
- **Bug fixes need a regression test.** Write a test that fails before your fix and passes after.
- **Tests should be deterministic.** Use `t.TempDir()` for filesystem tests. Don't depend on network, time, or external state.
- **Match existing test conventions.** Look at `internal/store/store_test.go` or `cmd/pressie/cmd_test.go` for the pattern.

## Code Style

- **Follow existing patterns.** The codebase is small — read the existing code before adding new abstractions.
- **No external dependencies unless necessary.** The project currently has zero external deps. Prefer the standard library.
- **Keep it boring.** No clever tricks, no premature abstraction. The code should be readable in six months.
- **Atomic writes for filesystem changes.** If you add a `Save*` function, use the `writeAtomic` pattern in `internal/store/store.go`.
- **One concern per change.** Don't mix refactors with feature additions in the same PR.

## Architecture

```
cmd/pressie/          CLI entry point and command handlers
internal/config/      Argument parsing
internal/store/       JSON file read/write (Gift, Idea, ContactFile, IndexFile)
internal/gifts/       Contact resolution (slug, fuzzy match, plugins)
internal/contacts/     Plugin execution
internal/web/          Web server (HTTP API + embedded SPA)
docs/                 Design docs and specifications
plugins/              Example contact plugins
```

The CLI and web server share the same data layer (`internal/store` and `internal/gifts`). No separate API — both call the same Go packages directly.

## Pull Request Checklist

Before opening a PR:

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] New features have tests
- [ ] Bug fixes have regression tests
- [ ] Commit messages describe what and why
- [ ] No unrelated changes in the PR
- [ ] You tested the behavior yourself (not just the AI's word)

## Reporting Bugs

Open a GitHub issue with:

1. What you did (exact commands or steps).
2. What you expected.
3. What happened instead.
4. The output of `pressie version`.

## Suggesting Features

Open a GitHub issue with the use case — not just the feature. "I want to track thank-you notes" is more useful than "add a thanked field." The design docs in `docs/` capture the project's direction; check there first.

## License

By contributing, you agree that your contributions are licensed under the MIT license, same as the project.