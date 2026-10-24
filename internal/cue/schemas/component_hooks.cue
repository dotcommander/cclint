// ============================================================================
// Shared Component Hook Definitions
// ============================================================================
//
// Single owner of the hook contract shared by agent.cue, command.cue and
// skill.cue, which previously repeated these structures byte for byte.
//
// This file is a definitions-only fragment: it has NO package clause and is
// never published as a standalone schema root. The schema loader reads it once
// and appends these definitions to every root before compiling (see
// schema_loader.go), so each composition still has exactly one package
// clause. The component-specific names below remain the accepted surfaces so
// every component root keeps its existing public contract.
//
// Settings hooks intentionally keep a richer, different contract; they stay
// defined separately in settings.cue (#HookType/#HookCommand/#Hook).

// Component hook command definition: shell `command` form or exec `args` form.
#ComponentHookCommand: {
	type:     "command"
	command?: string             // shell form
	args?:    [...string]        // exec form (v2.1.139+), alternative to command
	timeout?: int                // timeout in seconds
	once?:    bool               // run only once per session (v2.1.0+)
	continueOnBlock?: bool       // PostToolUse only (v2.1.139+)
	onFailure?:         "continue" | "block" // hook failure policy: continue lets the action through, block stops it (v2.1.295+)
	"if"?: string                // conditional filter using permission rule syntax (v2.1.85+)
}

// Component hook entry - matcher is optional (e.g., Stop hook doesn't need a matcher)
#ComponentHook: {
	matcher?: string                  // optional tool matcher (e.g., "Bash", "Write")
	hooks: [...#ComponentHookCommand] // array of commands to run
}

// Component hooks configuration - maps event names to hook arrays.
// Known events: PreToolUse, PostToolUse, Stop, Start.
// Allow any string key for future event types.
#ComponentHooks: {
	[string]: [...#ComponentHook]
}

// Component-specific names remain aliases to the shared structures: no
// per-root struct copies, unchanged accepted surfaces.
#AgentHookCommand:   #ComponentHookCommand
#AgentHook:          #ComponentHook
#AgentHooks:         #ComponentHooks

#CommandHookCommand: #ComponentHookCommand
#CommandHook:        #ComponentHook
#CommandHooks:       #ComponentHooks

#SkillHookCommand:   #ComponentHookCommand
#SkillHook:          #ComponentHook
#SkillHooks:         #ComponentHooks
