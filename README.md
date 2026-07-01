# agent-deck

`deck` is a task-first CLI for tracking agent work.

The core object is a task, not a terminal tab or an agent session. A task keeps
the goal, status, context, agent runs, artifacts, and the next human attention
point in one local JSON store.

## Install

```sh
go install github.com/sorafujitani/agent-deck/cmd/deck@latest
```

For local development:

```sh
go test ./...
go run ./cmd/deck help
```

## Architecture

The CLI is wired with manual constructor injection:

- `cmd/deck` is the composition root.
- `internal/cli` parses arguments and handles stdout/stderr.
- `internal/deck` owns the task model, lifecycle rules, and application service.
- `internal/storage` persists the deck state as JSON.

This keeps the domain logic testable without touching the filesystem, while the
CLI and storage adapters remain small and replaceable.

## Store

By default, `deck` stores data under the OS user config directory:

```sh
deck path
```

For tests or isolated workspaces, set `AGENT_DECK_HOME`:

```sh
AGENT_DECK_HOME=/tmp/agent-deck deck inbox
```

## Usage

Create a task:

```sh
deck new "Review PR #123" \
  --repo /path/to/repo \
  --pr https://github.com/owner/repo/pull/123 \
  --context "focus on API behavior" \
  --next "read the diff"
```

List open tasks in attention order:

```sh
deck inbox
deck inbox --json
deck inbox --repo /path/to/repo
```

Show the full case file for a task. Without an ID, `show` resolves the top
attention item from `deck inbox`:

```sh
deck show
deck show latest --json
deck show --repo /path/to/repo
deck show tsk_1234abcd
```

Track an agent run:

```sh
deck run latest --agent codex --summary "Reviewed diff and found no blockers"
deck run --repo /path/to/repo --agent codex --summary "Reviewed diff and found no blockers"
deck run tsk_1234abcd --agent codex --summary "Reviewed diff and found no blockers"
```

Attach an artifact:

```sh
deck artifact ./review.md --kind report --note "review output"
deck artifact --repo /path/to/repo ./review.md --kind report --note "review output"
deck artifact tsk_1234abcd ./review.md --kind report --note "review output"
```

Update status or next action:

```sh
deck update --status blocked --next "waiting for CI logs"
deck update --repo /path/to/repo --status blocked --next "waiting for CI logs"
deck update tsk_1234abcd --status blocked --next "waiting for CI logs"
```

Mark a task done:

```sh
deck done
deck done --repo /path/to/repo
deck done tsk_1234abcd
```

## Statuses

- `inbox`
- `ready`
- `running`
- `needs-review`
- `blocked`
- `failed`
- `done`
