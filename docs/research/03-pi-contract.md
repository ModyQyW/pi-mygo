# The pi contract we depend on

Everything below was read from the installed pi 1.5.2 documentation and confirmed
against the live binary where noted. This is the surface the app is allowed to
use; anything not listed here is not a dependency, and therefore cannot break us.

## Process

```sh
pi --mode rpc [--session <path|id>] [--session-dir <dir>] [--name <name>]
              [--provider <p>] [--model <m>] [--thinking <level>]
              [--extension <path>] [--no-extensions] [--no-session]
              [--no-context-files] [--approve|--no-approve] [--offline]
```

Working directory determines the project. RPC rejects `@file` prompt arguments;
send files as text or images. Startup cost measured locally: ~0.4 s before the
first record.

Framing rules we must obey:

- one JSON object per line, terminated by `\n`, on stdin and stdout;
- split stdin/stdout **only** on LF; `U+2028`/`U+2029` are legal inside strings
  and Node's `readline` gets this wrong, so we do our own splitter;
- strip one optional `\r` before parsing;
- drain stdout continuously; a stalled reader stalls pi;
- stdout carries protocol records only; diagnostics are on stderr and are never
  parsed as protocol.

Shutdown: close stdin, keep draining stdout until exit, wait, then stop reading.
A 5 s grace period, then `SIGTERM`, then `SIGKILL`. Violating the order crashes
pi with an unhandled `EPIPE`.

## Commands the app uses

| Command | Purpose | Notes |
| --- | --- | --- |
| `get_state` | model, thinking level, streaming, compacting, session file/id/name, queue counts | first call after spawn |
| `get_messages` | full linear conversation | opened on session restore |
| `get_entries` / `get_entries {since}` | append-order entries incl. abandoned branches | durable cursor by entry id |
| `get_tree` | full tree with labels | tree drawer |
| `prompt` | send a message | `disposition`: `started`, `queued`, `handled` |
| `steer` / `follow_up` | queue while running | `queued` or `handled` |
| `abort` | stop the run | waits for idle before responding |
| `clear_queue` | drop queues, return their text | Esc behaviour |
| `set_model` / `get_available_models` | model selection | `set_model` fails for unknown ids |
| `set_thinking_level` / `get_available_thinking_levels` | thinking | levels differ per model |
| `compact` / `set_auto_compaction` | compaction | `compact` returns the summary |
| `set_auto_retry` / `abort_retry` | retry control | exposed in settings/diagnostics |
| `new_session` / `switch_session` / `fork` / `clone` | session lifecycle | all can be cancelled by an extension |
| `get_fork_messages` | forkable user messages | used by the fork menu |
| `set_session_name` | rename | mirrors to `session_info` entry |
| `export_html` | export | produces a self-contained HTML file |
| `get_session_stats` | tokens, cost, context usage | status row |
| `get_commands` | extension commands, prompt templates, skills | slash menu; built-ins excluded |
| `bash` / `abort_bash` | run a shell command in the session | not used in v0.1 |

Every command accepts an optional `id`; the response repeats it. Handling is
asynchronous, so correlate by id, never by order.

### `prompt` details that bite

- While streaming, `prompt` **must** carry `streamingBehavior` (`steer` or
  `followUp`) or it errors.
- `images` is an array of `{type:"image", data:<base64>, mimeType}`.
- `disposition: "handled"` means an extension consumed the prompt and no run
  started; do not wait for `agent_settled`.
- Success means accepted, not completed. Failures after acceptance arrive as
  events, never as a second response for that id.

## Events

```text
agent_start, agent_end{messages,willRetry}, agent_settled{aborted}
turn_start, turn_end{message,toolResults}
message_start{message}, message_update{usage,assistantMessageEvent}, message_end{message}
tool_execution_start{toolCallId,toolName,args}
tool_execution_update{toolCallId,toolName,args,partialResult}
tool_execution_end{toolCallId,toolName,result,isError,durationMs}
queue_update{steering,followUp}
entry_appended{entry}, session_info_changed{name}, thinking_level_changed{level}
compaction_start{reason}, compaction_end{reason,result?,aborted,errorMessage?,willRetry?}
auto_retry_start{attempt,maxAttempts,delayMs,errorMessage}, auto_retry_end{success,attempt,finalError?}
summarization_retry_scheduled/attempt_start/finished
bash_execution_update{id,delta}      (RPC only)
extension_error{extensionPath,event,error}
```

`message_update` is delta-only on the wire. Reconstruct with `contentIndex` and
replace reconstructed text at `text_end` / `thinking_end` / `toolcall_end`, and
the whole message at `message_end`. Never trust an accumulated buffer over the
final message.

`agent_end` is not the end of work — retries, compaction recovery, steering and
follow-ups can continue. Only `agent_settled` means "no automatic work left".

## Extension UI subprotocol

Requests: `{type:"extension_ui_request", id, method, ...}`.

| Method | Expects response | v0.1 handling |
| --- | --- | --- |
| `select` | `{value}` or `{cancelled:true}` | inline picker in the transcript |
| `confirm` | `{confirmed}` or `{cancelled}` | inline confirm card |
| `input` | `{value}` or `{cancelled}` | inline single-line field |
| `editor` | `{value}` or `{cancelled}` | inline multiline editor |
| `notify` | none | toast, `notifyType` info/warning/error |
| `setStatus` | none | status chip, ANSI stripped, cleared when `statusText` absent |
| `setWidget` | none | ignored, counted in diagnostics |
| `setTitle` | none | applied to the window title |
| `set_editor_text` | none | inserted into the composer |

Measured on the real installation: five fire-and-forget records arrive
immediately after spawn from the user's own extensions, with ANSI escapes in
`statusText`. They must never block startup.

## Message shapes

`UserMessage.content` is a string **or** an array of text/image blocks.
`AssistantMessage.content` is an array of text/thinking/toolCall blocks plus
`usage`, `model`, `provider`, `stopReason`, `errorMessage`.
`ToolResultMessage` carries `toolCallId`, `toolName`, `content`, `isError`,
optional `details` and `usage`.
Coding-agent extras: `bashExecution`, `custom`, `branchSummary`,
`compactionSummary`. Unknown roles must be tolerated and rendered as a generic
notice rather than dropped.

`Usage` is `{input, output, cacheRead, cacheWrite, totalTokens, cost{...}}`.
`reasoning` is already included in `output`; never add it twice.

## Session files

Read-only for us:

```text
~/.pi/agent/sessions/--<escaped-cwd>--/<timestamp>_<session-id>.jsonl
```

Header: `{type:"session", version:3, id, timestamp, cwd, parentSession?}`.
Entries are `message`, `model_change`, `thinking_level_change`, `usage`,
`compaction`, `branch_summary`, `label`, `session_info`, `custom_message`,
`context_edit`. Tree via `id`/`parentId`; `session_info.name` is the display
name; `label.targetId` is a bookmark.

The project list is built from headers (`cwd`) plus `session_info` names, with
the small JSONL tail read for the last message when a preview is needed. Nothing
is written by the app; deletion goes through `trash`, matching pi's own
behaviour.

## Authentication

```sh
pi auth check --provider <p> --json     # 0 ready, 1 not_ready, 2 invalid
pi auth login <provider> [--api-key]    # interactive terminal
pi auth logout <provider>
```

The app never reads or writes `~/.pi/agent/auth.json`. Login runs in the system
terminal, and the app polls `pi auth check`.

## Deliberately unused

`get_messages` on every turn (we use events after the first load), `bash` (the
composer is not a terminal), SDK/SDK-only APIs, any internal TypeScript import,
and pi's undocumented HTTP surface. The bridge extension mechanism exists but is
unused until a real RPC gap blocks a v0.1 feature.
