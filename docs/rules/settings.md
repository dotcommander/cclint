# Settings Lint Rules

Rules enforced when linting `settings.json` files.

---

## Hook Structure Rules (048-057)

### Rule 048: JSON Parse Error

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Settings file must contain valid JSON. This rule catches syntax errors in JSON parsing.

**Pass Criteria:**
File contains well-formed JSON with proper syntax (valid braces, quotes, commas, etc.)

**Fail Message:**
`Error parsing JSON: [error details]`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Settings file must contain valid JSON structure"

---

### Rule 049: Unknown Hook Event Name

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Hook event names must match Anthropic's documented event types. Valid events: PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, Notification, UserPromptSubmit, Stop, SubagentStart, SubagentStop, PreCompact, SessionStart, SessionEnd, TeammateIdle, TaskCompleted, ConfigChange, WorktreeCreate, WorktreeRemove.

**Pass Criteria:**
Every key in the `hooks` object matches a valid hook event name from Anthropic documentation.

**Fail Message:**
`Unknown hook event '[eventName]'. Valid events: PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, Notification, UserPromptSubmit, Stop, SubagentStart, SubagentStop, PreCompact, SessionStart, SessionEnd, TeammateIdle, TaskCompleted, ConfigChange, WorktreeCreate, WorktreeRemove`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Claude Code provides 17 hook events that execute at different lifecycle points: PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, Notification, UserPromptSubmit, Stop, SubagentStart, SubagentStop, PreCompact, SessionStart, SessionEnd, TeammateIdle, TaskCompleted, ConfigChange, WorktreeCreate, WorktreeRemove"

---

### Rule 050: Hook Configuration Not Array

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook event's configuration must be an array of hook matchers. The structure is `{ "EventName": [...] }` where the value is an array.

**Pass Criteria:**
The value for each hook event key is a JSON array.

**Fail Message:**
`Event '[eventName]': hook configuration must be an array`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Hook event structure: `{ 'EventName': [...] }` where the value is an array"

---

### Rule 051: Hook Matcher Not Object

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each element in a hook event's array must be an object containing `matcher` and `hooks` fields.

**Pass Criteria:**
Each array element is a JSON object (not a string, number, or array).

**Fail Message:**
`Event '[eventName]' hook [index]: must be an object with 'matcher' and 'hooks' fields`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Each element in a hook event's array must be an object containing 'matcher' and 'hooks' fields"

---

### Rule 052: Missing 'matcher' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook matcher object must contain a `matcher` field that defines when the hook should trigger.

**Pass Criteria:**
Hook matcher object contains the required `matcher` field.

**Fail Message:**
`Event '[eventName]' hook [index]: missing required field 'matcher'`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Each hook matcher object must contain a 'matcher' field that defines when the hook should trigger"

---

### Rule 053: Missing 'hooks' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook matcher object must contain a `hooks` field that defines the array of hooks to execute when the matcher matches.

**Pass Criteria:**
Hook matcher object contains the required `hooks` field.

**Fail Message:**
`Event '[eventName]' hook [index]: missing required field 'hooks'`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Each hook matcher object must contain a 'hooks' field that defines the array of hooks to execute"

---

### Rule 054: Inner Hooks Not Array

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
The `hooks` field within a matcher must be an array of hook objects.

**Pass Criteria:**
The `hooks` field value is a JSON array.

**Fail Message:**
`Event '[eventName]' hook [index]: 'hooks' field must be an array`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "The 'hooks' field within a matcher must be an array of hook objects"

---

### Rule 055: Hook Object Not Object

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each individual hook within the `hooks` array must be an object containing hook configuration (type, command/prompt, etc.).

**Pass Criteria:**
Each element in the inner `hooks` array is a JSON object.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: must be an object`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Each individual hook within the 'hooks' array must be an object containing hook configuration"

---

### Rule 056: Missing 'type' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook object must specify its `type` field (either "command" or "prompt").

**Pass Criteria:**
Hook object contains the required `type` field.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: missing required field 'type'`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Each hook object must specify its 'type' field (either 'command' or 'prompt')"

---

### Rule 057: Type Not String

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
The `type` field must be a string value, not a number, boolean, or object.

**Pass Criteria:**
The `type` field contains a JSON string value.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: 'type' must be a string`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "The 'type' field must be a string value"

---

## Hook Type Rules (058-061)

### Rule 058: Invalid Hook Type

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Hook type must be one of the valid types. As of Claude Code 2.1.0, plugins can use "prompt" and "agent" types in addition to "command". As of Claude Code 2.1.63, "http" is also a valid type.

**Pass Criteria:**
The `type` field value is exactly "command", "prompt", "agent", or "http".

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: invalid type '[hookType]'. Valid types: command, prompt, agent, http`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - Hook type values (command, prompt, agent, http)

---

### Rule 059: Command Type Missing 'command' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
When hook type is "command", the hook object must include a `command` field containing the shell command to execute.

**Pass Criteria:**
Hook object with `"type": "command"` contains a `command` field.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: type 'command' requires 'command' field`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "When hook type is 'command', the hook object must include a 'command' field containing the shell command to execute"

---

### Rule 060: Prompt Type on Unsupported Event

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Prompt hooks are only supported for specific events: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest. Other events cannot use prompt-type hooks.

**Pass Criteria:**
If hook type is "prompt", the parent event must be one of the supported prompt hook events.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: event '[eventName]' does not support prompt hooks. Prompt hooks only supported for: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Prompt hooks are supported for: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest"

---

### Rule 061: Prompt Type Missing 'prompt' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
When hook type is "prompt", the hook object must include a `prompt` field containing the prompt text to inject.

**Pass Criteria:**
Hook object with `"type": "prompt"` contains a `prompt` field.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: type 'prompt' requires 'prompt' field`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "When hook type is 'prompt', the hook object must include a 'prompt' field containing the prompt text to inject"

---

## Hook Security Rules (062-074)

### Rule 062: Unquoted Variable Expansion

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Variable expansions ($VAR or ${VAR}) should be quoted to prevent word splitting and pathname expansion. Unquoted variables can lead to unexpected behavior or security issues if they contain spaces or special characters.

**Pass Criteria:**
All variable references in command strings are quoted: `"$VAR"` or `'$VAR'`.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Unquoted variable expansion detected. Use "$VAR" to prevent word splitting`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Always quote shell variables - Use `\"$VAR\"` not `$VAR`"

---

### Rule 063: Path Traversal Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `..` pattern in paths can be used for path traversal attacks, accessing files outside intended directories. This is a potential security risk.

**Pass Criteria:**
Command string does not contain `..` sequences.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Path traversal '..' detected in hook command - potential security risk`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Block path traversal - Check for `..` in file paths"

---

### Rule 064: Hardcoded Absolute Path

**Severity:** warning
**Component:** settings
**Category:** best-practice

**Description:**
Hardcoded absolute paths (like `/Users/username/project` or `/home/user/file`) make hooks non-portable. Use `$CLAUDE_PROJECT_DIR` for project-relative paths to ensure hooks work across machines.

**Pass Criteria:**
If absolute paths are present (starting with /Users, /home, /var, /tmp, /etc), the command also uses `$CLAUDE_PROJECT_DIR` or the path is legitimately system-wide.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Hardcoded absolute path detected. Consider using $CLAUDE_PROJECT_DIR for portability`

**Source:** cclint observation - Portability best practice: Use `$CLAUDE_PROJECT_DIR` for project-relative paths to ensure hooks work across machines and environments

---

### Rule 065: .env File Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing `.env` files in hooks may expose secrets. Ensure that hook commands do not log or transmit environment secrets.

**Pass Criteria:**
Command does not reference `.env` files, or access is verified to be secure.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing .env file - ensure secrets are not logged`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Skip sensitive files - Avoid `.env`, API keys, credentials"

---

### Rule 066: .git Directory Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing the `.git` directory can expose repository internals, credentials, or sensitive history. This is a potential security concern.

**Pass Criteria:**
Command does not access `.git/` paths.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing .git directory - potential security concern`

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - "Skip sensitive files - Avoid `.git/`, credentials"

---

### Rule 067: Credentials File Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing files with "credentials" in the name suggests handling of sensitive authentication data. Ensure secure handling and no exposure of secrets.

**Pass Criteria:**
Command does not reference files containing the word "credentials".

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing credentials file - ensure secure handling`

**Source:** cclint observation - Security best practice: Accessing files with "credentials" in the name suggests handling of sensitive authentication data that should never be logged or transmitted

---

### Rule 068: .ssh Directory Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `.ssh` directory contains private keys and SSH configuration. Accessing it in hooks is a high security risk and should be carefully reviewed.

**Pass Criteria:**
Command does not access `.ssh/` paths.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing .ssh directory - high security risk`

**Source:** cclint observation - Security best practice: The `.ssh` directory contains private keys and SSH configuration that should never be exposed, logged, or transmitted

---

### Rule 069: .aws Directory Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `.aws` directory contains AWS credentials and configuration. Accessing it in hooks can expose cloud credentials. Ensure no secrets are logged or transmitted.

**Pass Criteria:**
Command does not access `.aws/` paths.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing AWS config directory - ensure no secrets exposed`

**Source:** cclint observation - Security best practice: The `.aws` directory contains AWS credentials and configuration that can expose cloud credentials if logged or transmitted

---

### Rule 070: SSH Private Key Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing SSH private key files (id_rsa, id_ed25519, id_dsa) in hooks is a high security risk. Private keys should never be logged, transmitted, or exposed.

**Pass Criteria:**
Command does not reference SSH private key filenames.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing SSH private key - high security risk`

**Source:** cclint observation - Security best practice: SSH private key files (id_rsa, id_ed25519, id_dsa) should never be logged, transmitted, or exposed in any way

---

### Rule 071: eval Command Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `eval` command executes arbitrary strings as shell commands, creating a potential command injection vulnerability if user input or external data is involved.

**Pass Criteria:**
Command does not use the `eval` keyword.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: eval command detected - potential command injection risk`

**Source:** cclint observation - Security best practice: The `eval` command executes arbitrary strings as shell commands, creating command injection vulnerabilities if user input or external data is involved

---

### Rule 072: Command Substitution Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Command substitution using `$(...)` can execute arbitrary commands. Ensure that any substituted commands do not include unsanitized input to prevent command injection.

**Pass Criteria:**
Command does not use `$(...)` syntax, or usage is verified to be safe.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Command substitution detected - ensure input is sanitized`

**Source:** cclint observation - Security best practice: Command substitution using `$(...)` can execute arbitrary commands and must not include unsanitized input to prevent command injection

---

### Rule 073: Backtick Substitution Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Backtick command substitution (`` `command` ``) can execute arbitrary commands. This is an older shell syntax with the same risks as `$(...)`. Ensure input is sanitized.

**Pass Criteria:**
Command does not use backtick (`` ` ``) command substitution, or usage is verified to be safe.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Backtick command substitution detected - ensure input is sanitized`

**Source:** cclint observation - Security best practice: Backtick command substitution (`` `command` ``) can execute arbitrary commands with the same risks as `$(...)` and must have sanitized input

---

### Rule 074: Redirecting to /dev/

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Redirecting output to `/dev/` paths (like `/dev/null`, `/dev/tcp/`, or device files) should be reviewed to ensure it's intentional and not masking malicious activity or causing unintended side effects.

**Pass Criteria:**
Command does not redirect to `/dev/` paths, or redirection is verified to be intentional.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Redirecting to /dev/ - verify this is intentional`

**Source:** cclint observation - Security best practice: Redirecting output to `/dev/` paths (like `/dev/null`, `/dev/tcp/`) should be reviewed to ensure it's intentional and not masking malicious activity

---

## New Settings Fields (v2.1.0+)

Claude Code 2.1.0 introduced new settings.json fields:

| Field | Type | Description |
|-------|------|-------------|
| `language` | string | Configure Claude's response language (e.g., "japanese") |
| `respectGitignore` | bool | Control @-mention file picker behavior |

### New Hook Fields

| Field | Type | Description |
|-------|------|-------------|
| `once` | bool | Run hook only once per session |

### New Hook Types (Plugins)

| Type | Description |
|------|-------------|
| `command` | Execute a shell command (original) |
| `prompt` | Modify Claude's prompt (plugins, v2.1.0+) |
| `agent` | Invoke an agent (plugins, v2.1.0+) |
| `http` | Make an HTTP request (v2.1.63+) |

### Example settings.json

```json
{
  "language": "japanese",
  "respectGitignore": true,
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Session started'",
            "once": true
          }
        ]
      }
    ]
  }
}
```

---

## Summary

Settings linting enforces:
- **Structural rules (048-057)**: Validate JSON structure and hook schema conformance
- **Type rules (058-061)**: Ensure hook types are valid and have required fields
- **Security rules (062-074)**: Detect common security anti-patterns in hook commands

All rules marked with `cue.SourceAnthropicDocs` validate against Anthropic's official hook specification.
All rules marked with `cue.SourceCClintObserve` are security best practices derived from common vulnerabilities.
