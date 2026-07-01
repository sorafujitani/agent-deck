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
```

Show the full case file for a task:

```sh
deck show tsk_1234abcd
```

Track an agent run:

```sh
deck run tsk_1234abcd --agent codex --summary "Reviewed diff and found no blockers"
```

Attach an artifact:

```sh
deck artifact tsk_1234abcd ./review.md --kind report --note "review output"
```

Update status or next action:

```sh
deck update tsk_1234abcd --status blocked --next "waiting for CI logs"
```

Mark a task done:

```sh
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
