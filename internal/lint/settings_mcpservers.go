package lint

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/cue"
)

// validateMCPServers validates the mcpServers configuration map.
// Each entry maps a server name to an object with command, args, env, and cwd fields.
func validateMCPServers(mcpServers any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	serversMap, ok := mcpServers.(map[string]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "mcpServers must be an object mapping server names to configurations",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for serverName, serverConfig := range serversMap {
		if serverName == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "mcpServers: server name must not be empty",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		serverMap, ok := serverConfig.(map[string]any)
		if !ok {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': server configuration must be an object", serverName),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		errors = append(errors, validateMCPServerEntry(serverName, serverMap, filePath)...)
	}

	return errors
}

// validateMCPServerEntry validates a single MCP server configuration entry.
func validateMCPServerEntry(serverName string, serverMap map[string]any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Validate command field (required, non-empty string)
	cmdVal, cmdExists := serverMap["command"]
	if !cmdExists {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': missing required field 'command'", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	} else if cmdStr, ok := cmdVal.(string); !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'command' must be a string", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	} else if cmdStr == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'command' must not be empty", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	// Validate args field (optional, must be array of strings)
	if argsVal, argsExists := serverMap["args"]; argsExists {
		errors = append(errors, validateMCPServerArgs(serverName, argsVal, filePath)...)
	}

	// Validate env field (optional, must be object with string values)
	if envVal, envExists := serverMap["env"]; envExists {
		errors = append(errors, validateMCPServerEnv(serverName, envVal, filePath)...)
	}

	// Validate cwd field (optional, must be a string)
	if cwdVal, cwdExists := serverMap["cwd"]; cwdExists {
		if _, ok := cwdVal.(string); !ok {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': 'cwd' must be a string", serverName),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}

	return errors
}

// validateMCPServerArgs validates the args field of an MCP server entry.
func validateMCPServerArgs(serverName string, argsVal any, filePath string) []cue.ValidationError {
	argsArray, ok := argsVal.([]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'args' must be an array of strings", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}
	var errors []cue.ValidationError
	for i, arg := range argsArray {
		if _, isStr := arg.(string); !isStr {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': args[%d] must be a string", serverName, i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}
	return errors
}

// validateMCPServerEnv validates the env field of an MCP server entry.
func validateMCPServerEnv(serverName string, envVal any, filePath string) []cue.ValidationError {
	envMap, ok := envVal.(map[string]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'env' must be an object with string values", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}
	var errors []cue.ValidationError
	for envKey, envValue := range envMap {
		if _, isStr := envValue.(string); !isStr {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': env '%s' value must be a string", serverName, envKey),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}
	return errors
}
