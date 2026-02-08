package cli

import (
	"fmt"
	"regexp"
	"strings"
)

// Cycle represents a circular dependency in the component graph
type Cycle struct {
	Path []string // Component names in cycle order (last element == first element)
	Type string   // "command-agent-command", "agent-skill-agent", etc.
}

// getNeighbors returns the neighboring nodes for a given component in the dependency graph.
// This extracts neighbor resolution logic from DetectCycles for testability.
func (v *CrossFileValidator) getNeighbors(componentType, name string) []string {
	var neighbors []string
	taskPattern := regexp.MustCompile(`Task\(([^,\)]+)`)

	switch componentType {
	case "command":
		if cmd, exists := v.commands[name]; exists {
			neighbors = v.extractAgentRefsFromTask(cmd.Contents, taskPattern, "")
		}
	case "agent":
		if agent, exists := v.agents[name]; exists {
			// Agents reference skills
			for _, skillRef := range findSkillReferences(agent.Contents) {
				if _, exists := v.skills[skillRef]; exists {
					neighbors = append(neighbors, "skill:"+skillRef)
				}
			}
			// Agents might delegate to other agents (exclude self)
			neighbors = append(neighbors, v.extractAgentRefsFromTask(agent.Contents, taskPattern, name)...)
		}
	case "skill":
		if skill, exists := v.skills[name]; exists {
			// Skills might delegate to agents
			neighbors = append(neighbors, v.extractAgentRefsFromPatterns(skill.Contents)...)
			// Skills might reference other skills (exclude self)
			for _, skillRef := range findSkillReferences(skill.Contents) {
				if skillRef != name {
					if _, exists := v.skills[skillRef]; exists {
						neighbors = append(neighbors, "skill:"+skillRef)
					}
				}
			}
		}
	}
	return neighbors
}

// extractAgentRefsFromTask extracts agent references from Task() calls.
// excludeName is used to prevent self-references (pass "" to include all).
func (v *CrossFileValidator) extractAgentRefsFromTask(contents string, pattern *regexp.Regexp, excludeName string) []string {
	var refs []string
	matches := pattern.FindAllStringSubmatch(contents, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			agentRef := strings.TrimSpace(match[1])
			agentRef = strings.Trim(agentRef, `"'`)
			if strings.Contains(agentRef, "subagent_type") {
				continue
			}
			if excludeName != "" && agentRef == excludeName {
				continue
			}
			if _, exists := v.agents[agentRef]; exists {
				refs = append(refs, "agent:"+agentRef)
			}
		}
	}
	return refs
}

// extractAgentRefsFromPatterns extracts agent references from skill content patterns.
func (v *CrossFileValidator) extractAgentRefsFromPatterns(contents string) []string {
	var refs []string
	patterns := []string{
		`delegate to\s+([a-z0-9][a-z0-9-]+)`,
		`use\s+([a-z0-9][a-z0-9-]+)`,
		`Task\(([a-z0-9][a-z0-9-]+)`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(contents, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				agentRef := strings.TrimSpace(match[1])
				if _, exists := v.agents[agentRef]; exists {
					refs = append(refs, "agent:"+agentRef)
				}
			}
		}
	}
	return refs
}

// DetectCycles finds circular dependencies using DFS with color marking.
// Returns all cycles found in the component graph.
//
// Algorithm: DFS with three colors:
//   - white (0): unvisited
//   - gray (1): currently visiting (in recursion stack)
//   - black (2): completely visited
//
// Back edges (gray -> gray) indicate cycles.
func (v *CrossFileValidator) DetectCycles() []Cycle {
	var cycles []Cycle

	// Track visit state: 0=white, 1=gray, 2=black
	visitState := make(map[string]int)

	// Track current path for cycle reconstruction
	path := []string{}
	inPath := make(map[string]bool)

	// DFS visit function
	var visit func(componentType, name string)
	visit = func(componentType, name string) {
		nodeID := componentType + ":" + name

		// Mark as gray (visiting)
		visitState[nodeID] = 1
		path = append(path, nodeID)
		inPath[nodeID] = true

		// Visit neighbors
		for _, neighbor := range v.getNeighbors(componentType, name) {
			switch {
			case visitState[neighbor] == 0:
				// White: unvisited, recurse
				parts := strings.SplitN(neighbor, ":", 2)
				if len(parts) == 2 {
					visit(parts[0], parts[1])
				}
			case visitState[neighbor] == 1 && inPath[neighbor]:
				// Gray and in current path: back edge = cycle detected
				if cyclePath := reconstructCycle(path, neighbor); cyclePath != nil {
					cycles = append(cycles, Cycle{
						Path: cyclePath,
						Type: determineCycleType(cyclePath),
					})
				}
				// Black (2): already visited, skip
			}
		}

		// Mark as black (visited)
		visitState[nodeID] = 2
		path = path[:len(path)-1]
		delete(inPath, nodeID)
	}

	// Start DFS from all commands (typical entry points)
	for cmdName := range v.commands {
		nodeID := "command:" + cmdName
		if visitState[nodeID] == 0 {
			visit("command", cmdName)
		}
	}

	// Also check agents not reachable from commands
	for agentName := range v.agents {
		nodeID := "agent:" + agentName
		if visitState[nodeID] == 0 {
			visit("agent", agentName)
		}
	}

	// Also check skills not reachable from agents/commands
	for skillName := range v.skills {
		nodeID := "skill:" + skillName
		if visitState[nodeID] == 0 {
			visit("skill", skillName)
		}
	}

	return cycles
}

// reconstructCycle extracts the cycle path from the DFS stack starting at the neighbor node.
// Returns nil if the neighbor is not found in the path.
func reconstructCycle(path []string, neighbor string) []string {
	cycleStart := -1
	for i, p := range path {
		if p == neighbor {
			cycleStart = i
			break
		}
	}
	if cycleStart < 0 {
		return nil
	}
	cyclePath := make([]string, len(path)-cycleStart+1)
	copy(cyclePath, path[cycleStart:])
	cyclePath[len(cyclePath)-1] = neighbor // Close the cycle
	return cyclePath
}

// determineCycleType classifies the cycle based on component types involved
func determineCycleType(path []string) string {
	if len(path) == 0 {
		return "unknown"
	}

	types := make([]string, 0, len(path))
	for _, node := range path {
		parts := strings.SplitN(node, ":", 2)
		if len(parts) == 2 {
			types = append(types, parts[0])
		}
	}

	return strings.Join(types, " \u2192 ")
}

// FormatCycle formats a cycle for human-readable output
func FormatCycle(cycle Cycle) string {
	if len(cycle.Path) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, node := range cycle.Path {
		parts := strings.SplitN(node, ":", 2)
		if len(parts) == 2 {
			sb.WriteString(parts[1]) // Just the name
			if i < len(cycle.Path)-1 {
				sb.WriteString(" \u2192 ")
			}
		}
	}

	return sb.String()
}

// ChainLink represents a component in the delegation chain
type ChainLink struct {
	Type     string // "command", "agent", "skill"
	Name     string
	Path     string
	Lines    int
	Children []ChainLink
}

// TraceChain traces the full delegation chain starting from a component
func (v *CrossFileValidator) TraceChain(componentType string, name string) *ChainLink {
	switch componentType {
	case "command":
		return v.traceFromCommand(name)
	case "agent":
		return v.traceFromAgent(name)
	case "skill":
		return v.traceFromSkill(name)
	}
	return nil
}

func (v *CrossFileValidator) traceFromCommand(name string) *ChainLink {
	file, exists := v.commands[name]
	if !exists {
		return nil
	}

	link := &ChainLink{
		Type:  "command",
		Name:  name,
		Path:  file.RelPath,
		Lines: strings.Count(file.Contents, "\n") + 1,
	}

	// Find Task() delegations
	taskPattern := regexp.MustCompile(`Task\(([^,\)]+)`)
	matches := taskPattern.FindAllStringSubmatch(file.Contents, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			agentRef := strings.TrimSpace(match[1])
			agentRef = strings.Trim(agentRef, `"'`)
			if !strings.Contains(agentRef, "subagent_type") {
				if child := v.traceFromAgent(agentRef); child != nil {
					link.Children = append(link.Children, *child)
				}
			}
		}
	}

	return link
}

func (v *CrossFileValidator) traceFromAgent(name string) *ChainLink {
	file, exists := v.agents[name]
	if !exists {
		return nil
	}

	link := &ChainLink{
		Type:  "agent",
		Name:  name,
		Path:  file.RelPath,
		Lines: strings.Count(file.Contents, "\n") + 1,
	}

	// Find Skill references using comprehensive pattern matching
	skillRefs := findSkillReferences(file.Contents)
	for _, skillRef := range skillRefs {
		if child := v.traceFromSkill(skillRef); child != nil {
			link.Children = append(link.Children, *child)
		}
	}

	return link
}

func (v *CrossFileValidator) traceFromSkill(name string) *ChainLink {
	file, exists := v.skills[name]
	if !exists {
		return nil
	}

	return &ChainLink{
		Type:  "skill",
		Name:  name,
		Path:  file.RelPath,
		Lines: strings.Count(file.Contents, "\n") + 1,
	}
}

// FormatChain formats a chain link as a tree string
func FormatChain(link *ChainLink, indent string) string {
	if link == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s%s (%s, %d lines)\n", indent, link.Name, link.Type, link.Lines))

	for i, child := range link.Children {
		prefix := "├── "
		childIndent := indent + "│   "
		if i == len(link.Children)-1 {
			prefix = "└── "
			childIndent = indent + "    "
		}
		sb.WriteString(fmt.Sprintf("%s%s", indent, prefix))
		sb.WriteString(FormatChain(&child, childIndent)[len(childIndent):])
	}

	return sb.String()
}
