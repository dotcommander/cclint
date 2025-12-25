# Plugin Lint Rules

Documentation for all plugin-specific validation rules enforced by cclint.

---

## Overview

Plugin manifests are JSON files that define custom plugins for Claude Code. These rules validate both required fields (per Anthropic documentation) and best practices for plugin quality and discoverability.

## Required Fields (Rules 075-083)

### Rule 075: Missing 'name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include a `name` field that identifies the plugin.

**Pass Criteria:**
- The `name` field exists
- The `name` field is not empty

**Fail Message:**
`Required field 'name' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The name field is required and must be a string identifying your plugin."

---

### Rule 076: Empty 'name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `name` field cannot be an empty string.

**Pass Criteria:**
- The `name` field contains at least one character

**Fail Message:**
`Required field 'name' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The name field is required and must be a string identifying your plugin."

---

### Rule 077: Missing 'description' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include a `description` field that explains what the plugin does.

**Pass Criteria:**
- The `description` field exists
- The `description` field is not empty

**Fail Message:**
`Required field 'description' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The description field is required and provides a brief explanation of what your plugin does."

---

### Rule 078: Empty 'description' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `description` field cannot be an empty string.

**Pass Criteria:**
- The `description` field contains at least one character

**Fail Message:**
`Required field 'description' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The description field is required and provides a brief explanation of what your plugin does."

---

### Rule 079: Missing 'version' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include a `version` field following semantic versioning.

**Pass Criteria:**
- The `version` field exists
- The `version` field is not empty

**Fail Message:**
`Required field 'version' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The version field is required and must follow semantic versioning (semver) format."

---

### Rule 080: Empty 'version' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `version` field cannot be an empty string.

**Pass Criteria:**
- The `version` field contains at least one character

**Fail Message:**
`Required field 'version' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The version field is required and must follow semantic versioning (semver) format."

---

### Rule 081: Missing 'author' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include an `author` object.

**Pass Criteria:**
- The `author` field exists as an object/map

**Fail Message:**
`Required field 'author' is missing`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The author field contains information about the plugin creator with required name subfield."

---

### Rule 082: Missing 'author.name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `author` object must include a `name` field identifying the plugin author.

**Pass Criteria:**
- The `author.name` field exists
- The `author.name` field is not empty

**Fail Message:**
`Required field 'author.name' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The author field contains information about the plugin creator with required name subfield."

---

### Rule 083: Empty 'author.name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `author.name` field cannot be an empty string.

**Pass Criteria:**
- The `author.name` field contains at least one character

**Fail Message:**
`Required field 'author.name' is missing or empty`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The author field contains information about the plugin creator with required name subfield."

---

## Constraints (Rules 084-087)

### Rule 084: Reserved word in name

**Severity:** error
**Component:** plugin
**Category:** security

**Description:**
Plugin names cannot use reserved words that conflict with core Claude Code namespaces.

**Pass Criteria:**
- The `name` field does not contain (case-insensitive):
  - `anthropic`
  - `claude`

**Fail Message:**
`Name '[name]' is a reserved word and cannot be used`

**Source:** [Anthropic Docs - Plugin Marketplaces](https://code.claude.com/docs/en/plugin-marketplaces) - "Reserved marketplace names include claude-code-marketplace, claude-code-plugins, claude-plugins-official, anthropic-marketplace, anthropic-plugins."

---

### Rule 085: Name exceeds 64 characters

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
Plugin names must not exceed 64 characters to ensure compatibility and usability.

**Pass Criteria:**
- The `name` field is 64 characters or fewer

**Fail Message:**
`Name exceeds 64 character limit ([N] chars)`

**Source:** cclint observation - Conservative limit inferred from tool name constraints, not explicitly documented in Anthropic plugin specification

---

### Rule 086: Description exceeds 1024 characters

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
Plugin descriptions must not exceed 1024 characters to ensure readability and UI compatibility.

**Pass Criteria:**
- The `description` field is 1024 characters or fewer

**Fail Message:**
`Description exceeds 1024 character limit ([N] chars)`

**Source:** cclint observation - Practical limit for UI display; official plugins range 47-174 characters, not explicitly documented in Anthropic specification

---

### Rule 087: Invalid semver version format

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `version` field must follow semantic versioning format (MAJOR.MINOR.PATCH).

**Pass Criteria:**
- The `version` matches the pattern: `^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`
- Examples: `1.0.0`, `2.3.4-beta.1`, `1.0.0+build.123`

**Fail Message:**
`Version '[version]' must be in semver format (e.g., 1.0.0)`

**Source:** [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference) - "The version field is required and must follow semantic versioning (semver) format."

---

## Best Practices (Rules 088-092)

### Rule 088: Missing 'homepage' field

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `homepage` field helps users discover more information about the plugin.

**Pass Criteria:**
- The `homepage` field exists with a valid URL

**Fail Message:**
`Consider adding 'homepage' field with project URL`

**Source:** cclint observation - Best practice for discoverability; field documented as optional in [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference)

---

### Rule 089: Missing 'repository' field

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `repository` field helps users find the source code and contribute to the plugin.

**Pass Criteria:**
- The `repository` field exists with a valid repository URL

**Fail Message:**
`Consider adding 'repository' field with source code URL`

**Source:** cclint observation - Best practice for open source contributions; field documented as optional in [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference)

---

### Rule 090: Missing 'license' field

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `license` field clarifies usage rights and legal terms for the plugin.

**Pass Criteria:**
- The `license` field exists with a valid SPDX license identifier

**Fail Message:**
`Consider adding 'license' field (e.g., MIT, Apache-2.0)`

**Source:** cclint observation - Best practice for legal clarity; field documented as optional in [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference)

---

### Rule 091: Missing 'keywords' array

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `keywords` array improves plugin discoverability in search and catalogs.

**Pass Criteria:**
- The `keywords` field exists as a non-empty array

**Fail Message:**
`Consider adding 'keywords' array for discoverability`

**Source:** cclint observation - Best practice for marketplace search; field documented as optional in [Anthropic Docs - Plugins Reference](https://code.claude.com/docs/en/plugins-reference)

---

### Rule 092: Description too short

**Severity:** suggestion
**Component:** plugin
**Category:** best-practice

**Description:**
Plugin descriptions should be at least 50 characters to provide adequate context for users.

**Pass Criteria:**
- The `description` field is 50 characters or longer

**Fail Message:**
`Description is only [N] chars - consider expanding for clarity`

**Source:** cclint observation - Best practice derived from analyzing official plugins (range 47-174 characters); ensures adequate user context

---

## Additional Validations

### JSON Parsing Error

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
Plugin manifest files must be valid JSON.

**Pass Criteria:**
- File can be parsed by Go's `encoding/json` package

**Fail Message:**
`Error parsing JSON: [error details]`

---

### Secrets Detection

**Severity:** warning
**Component:** plugin
**Category:** security

**Description:**
Plugin manifests are scanned for potentially exposed secrets (API keys, tokens, passwords).

**Pass Criteria:**
- No patterns matching common secret formats are detected

**Fail Message:**
Various messages depending on detected pattern type

**Source:** cclint observation - Security best practice using pattern detection for common secret formats

---

## Rule Categories Summary

| Category | Rule Range | Count | Severity |
|----------|------------|-------|----------|
| Required Fields | 075-083 | 9 | error |
| Constraints | 084-087 | 4 | error |
| Best Practices | 088-092 | 5 | suggestion |

**Total Plugin Rules:** 18

---

## Examples

### Valid Plugin Manifest

```json
{
  "name": "my-awesome-plugin",
  "description": "A comprehensive plugin that enhances Claude Code with custom functionality for advanced workflows",
  "version": "1.0.0",
  "author": {
    "name": "Jane Developer",
    "email": "jane@example.com"
  },
  "homepage": "https://github.com/janedev/my-awesome-plugin",
  "repository": "https://github.com/janedev/my-awesome-plugin",
  "license": "MIT",
  "keywords": ["productivity", "workflow", "automation"]
}
```

### Invalid Examples

#### Reserved Word Violation (Rule 084)
```json
{
  "name": "claude-helper",  // ❌ Contains reserved word "claude"
  "description": "A helper plugin",
  "version": "1.0.0",
  "author": {"name": "Developer"}
}
```

#### Name Too Long (Rule 085)
```json
{
  "name": "my-incredibly-long-plugin-name-that-exceeds-the-maximum-allowed-length-limit",  // ❌ 75 chars
  "description": "A plugin with a very long name",
  "version": "1.0.0",
  "author": {"name": "Developer"}
}
```

#### Invalid Semver (Rule 087)
```json
{
  "name": "my-plugin",
  "description": "A plugin with invalid version",
  "version": "v1.0",  // ❌ Not semver format
  "author": {"name": "Developer"}
}
```

#### Short Description (Rule 092)
```json
{
  "name": "my-plugin",
  "description": "A plugin",  // ⚠️ Only 8 chars (suggestion)
  "version": "1.0.0",
  "author": {"name": "Developer"}
}
```

---

## Implementation Details

### Source Attributions

Plugin validation rules come from two sources:

1. **Anthropic Docs** - Official requirements from Anthropic's plugin manifest specification
2. **cclint observe** - Best practices observed from high-quality plugin examples

### Validation Pipeline

```
JSON Parse → Required Fields → Constraints → Best Practices → Secrets Detection → Quality Scoring
```

### Related Commands

```bash
# Lint all plugin manifests
cclint plugins

# Show quality scores for plugins
cclint plugins --scores

# Show improvement recommendations
cclint plugins --improvements

# Verbose output with processing details
cclint plugins --verbose
```

---

## See Also

- [Agent Rules](agents.md)
- [Command Rules](commands.md)
- [Skill Rules](skills.md)
- [Quality Scoring](../scoring.md)
