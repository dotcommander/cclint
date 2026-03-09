package lint

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/cue"
)

// validateHooks validates hooks for settings (full event set)
func validateHooks(hooks any, filePath string) []cue.ValidationError {
	return validateHooksWithEvents(hooks, filePath, validHookEvents, eventLabel(validHookEvents))
}

// ValidateComponentHooks validates hooks for agents and skills (scoped event set)
func ValidateComponentHooks(hooks any, filePath string) []cue.ValidationError {
	return validateHooksWithEvents(hooks, filePath, validComponentHookEvents, eventLabel(validComponentHookEvents))
}

// validateHooksWithEvents validates the hooks section with specified allowed events
func validateHooksWithEvents(hooks any, filePath string, allowedEvents map[string]bool, eventLabel string) []cue.ValidationError {
	var errors []cue.ValidationError

	hooksMap, ok := hooks.(map[string]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "hooks must be an object mapping event names to hook configurations",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	// Validate each event name and its hooks
	for eventName, eventConfig := range hooksMap {
		errors = append(errors, validateHookEvent(eventName, eventConfig, filePath, allowedEvents, eventLabel)...)
	}

	return errors
}

func validateHookEvent(eventName string, eventConfig any, filePath string, allowedEvents map[string]bool, eventLabel string) []cue.ValidationError {
	if !allowedEvents[eventName] {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Unknown hook event '%s'. Valid events: %s", eventName, eventLabel),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	hookArray, ok := eventConfig.([]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s': hook configuration must be an array", eventName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	var errors []cue.ValidationError
	for i, hookMatcher := range hookArray {
		errors = append(errors, validateHookMatcher(hookMatcher, eventName, i, filePath)...)
	}

	return errors
}

// validateHookMatcher validates a single hook matcher entry within an event.
func validateHookMatcher(hookMatcher any, eventName string, idx int, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	hookMatcherMap, ok := hookMatcher.(map[string]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: must be an object with 'matcher' and 'hooks' fields", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	// Check for required 'matcher' field and validate toolName if present
	errors = append(errors, validateHookMatcherField(hookMatcherMap, eventName, idx, filePath)...)

	// Check for required 'hooks' field
	innerHooks, exists := hookMatcherMap["hooks"]
	if !exists {
		return append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: missing required field 'hooks'", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	innerHooksArray, ok := innerHooks.([]any)
	if !ok {
		return append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: 'hooks' field must be an array", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	for j, innerHook := range innerHooksArray {
		errors = append(errors, validateInnerHook(innerHook, eventName, idx, j, filePath)...)
	}

	return errors
}

// validateHookMatcherField validates the matcher field of a hook matcher entry.
func validateHookMatcherField(hookMatcherMap map[string]any, eventName string, idx int, filePath string) []cue.ValidationError {
	matcherVal, matcherExists := hookMatcherMap["matcher"]
	if !matcherExists {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: missing required field 'matcher'", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	matcherMap, isMap := matcherVal.(map[string]any)
	if !isMap {
		return nil
	}

	toolNameVal, exists := matcherMap["toolName"]
	if !exists {
		return nil
	}

	toolNameStr, isStr := toolNameVal.(string)
	if !isStr || toolNameStr == "" {
		return nil
	}

	location := fmt.Sprintf("Event '%s' hook %d matcher", eventName, idx)
	return validateMatcherToolName(toolNameStr, location, filePath)
}

// validateInnerHook validates a single inner hook entry (type, command/prompt fields).
func validateInnerHook(innerHook any, eventName string, hookIdx, innerIdx int, filePath string) []cue.ValidationError {
	innerHookMap, ok := innerHook.(map[string]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: must be an object", eventName, hookIdx, innerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	hookType, typeExists := innerHookMap["type"]
	if !typeExists {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: missing required field 'type'", eventName, hookIdx, innerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	hookTypeStr, ok := hookType.(string)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: 'type' must be a string", eventName, hookIdx, innerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	if !validHookTypes[hookTypeStr] {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: invalid type '%s'. Valid types: command, prompt, agent, http", eventName, hookIdx, innerIdx, hookTypeStr),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	hookCtx := hookContext{EventName: eventName, HookIdx: hookIdx, InnerIdx: innerIdx, FilePath: filePath}
	return validateInnerHookType(innerHookMap, hookTypeStr, hookCtx)
}

// hookContext holds context information for hook validation
type hookContext struct {
	EventName string
	HookIdx   int
	InnerIdx  int
	FilePath  string
}

// validateInnerHookType validates type-specific requirements for a hook entry.
func validateInnerHookType(hookMap map[string]any, hookType string, ctx hookContext) []cue.ValidationError {
	validator, ok := innerHookValidators[hookType]
	if !ok {
		return nil
	}
	return validator(hookMap, ctx)
}

type innerHookValidator func(map[string]any, hookContext) []cue.ValidationError

var innerHookValidators = map[string]innerHookValidator{
	cue.TypeCommand: validateCommandInnerHook,
	"prompt":        validatePromptInnerHook,
	cue.TypeHTTP:    validateHTTPInnerHook,
}

func validateCommandInnerHook(hookMap map[string]any, ctx hookContext) []cue.ValidationError {
	cmdVal, exists := hookMap["command"]
	if !exists {
		return []cue.ValidationError{{
			File:     ctx.FilePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: type 'command' requires 'command' field", ctx.EventName, ctx.HookIdx, ctx.InnerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	cmdStr, ok := cmdVal.(string)
	if !ok {
		return nil
	}

	return validateHookCommandSecurity(cmdStr, ctx)
}

func validatePromptInnerHook(hookMap map[string]any, ctx hookContext) []cue.ValidationError {
	var errors []cue.ValidationError
	if !promptHookEvents[ctx.EventName] {
		errors = append(errors, cue.ValidationError{
			File:     ctx.FilePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: event '%s' does not support prompt hooks. Prompt hooks only supported for: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest", ctx.EventName, ctx.HookIdx, ctx.InnerIdx, ctx.EventName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}
	if _, exists := hookMap["prompt"]; !exists {
		errors = append(errors, cue.ValidationError{
			File:     ctx.FilePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: type 'prompt' requires 'prompt' field", ctx.EventName, ctx.HookIdx, ctx.InnerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}
	return errors
}

func validateHTTPInnerHook(hookMap map[string]any, ctx hookContext) []cue.ValidationError {
	if _, exists := hookMap["url"]; exists {
		return nil
	}

	return []cue.ValidationError{{
		File:     ctx.FilePath,
		Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: type 'http' requires 'url' field", ctx.EventName, ctx.HookIdx, ctx.InnerIdx),
		Severity: "error",
		Source:   cue.SourceAnthropicDocs,
	}}
}
