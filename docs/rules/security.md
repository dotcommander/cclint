# Security Rules

Rules that detect potential security vulnerabilities and sensitive data exposure.

---

## Tool Validation

### Rule 093: Unknown tool in allowed-tools

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Validates that tool names in `allowed-tools` and `tools` frontmatter fields are recognized Claude Code tools. Prevents typos and ensures proper tool restrictions.

**Pass Criteria:**
- Tool names must match known tools: Read, Write, Edit, MultiEdit, Glob, Grep, LS, Bash, Task, WebFetch, WebSearch, AskUserQuestion, TodoWrite, Skill, LSP, NotebookEdit, EnterPlanMode, ExitPlanMode, KillShell, TaskOutput, or `*`
- Patterns like `Task(specialist-name)` and `Bash(npm:*)` are validated by base tool name
- Empty or whitespace-only tool names are ignored

**Fail Message:**
`Unknown tool '[tool-name]' in [field-name]. Check spelling or verify it's a valid tool.`

**Source:** cclint observation - Validates against [Anthropic Docs tool list](https://code.claude.com/docs/en/sub-agents)

---

## Secrets Detection

### Rule 094: Hardcoded API key pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects potential hardcoded API keys using common assignment patterns with key/value lengths typical of API credentials.

**Pass Criteria:**
- No matches for pattern: `(?i)(api[_-]?key|apikey)\s*[:=]\s*["'][^"']{10,}["']`
- API keys should be loaded from environment variables or secrets management

**Fail Message:**
`Possible hardcoded API key detected - use environment variables`

**Source:** [OWASP - Secrets in Source Code](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/01-Information_Gathering/05-Review_Webpage_Content_for_Information_Leakage)

---

### Rule 095: Hardcoded password pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects potential hardcoded passwords in variable assignments.

**Pass Criteria:**
- No matches for pattern: `(?i)(password|passwd|pwd)\s*[:=]\s*["'][^"']+["']`
- Passwords should be loaded from environment variables or secrets management

**Fail Message:**
`Possible hardcoded password detected - use secrets management`

**Source:** [CWE-798: Use of Hard-coded Credentials](https://cwe.mitre.org/data/definitions/798.html)

---

### Rule 096: Hardcoded secret/token pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects potential hardcoded secrets or tokens with sufficient length to be real credentials.

**Pass Criteria:**
- No matches for pattern: `(?i)(secret|token)\s*[:=]\s*["'][^"']{10,}["']`
- Secrets should be loaded from environment variables or secrets management

**Fail Message:**
`Possible hardcoded secret/token detected - use environment variables`

**Source:** [CWE-798: Use of Hard-coded Credentials](https://cwe.mitre.org/data/definitions/798.html)

---

### Rule 097: OpenAI API key pattern (sk-)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects OpenAI API keys by their distinctive `sk-` prefix format.

**Pass Criteria:**
- No matches for pattern: `sk-[a-zA-Z0-9]{20,}`
- Never commit API keys to version control

**Fail Message:**
`OpenAI API key pattern detected - never commit API keys`

**Source:** [OpenAI API Key Format](https://platform.openai.com/docs/api-reference/authentication) - sk- prefix pattern

---

### Rule 098: Slack bot token pattern (xoxb-)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects Slack bot tokens by their distinctive `xoxb-` prefix format.

**Pass Criteria:**
- No matches for pattern: `xoxb-[a-zA-Z0-9-]+`
- Use environment variables for Slack tokens

**Fail Message:**
`Slack bot token pattern detected - use environment variables`

**Source:** [Slack Token Types](https://api.slack.com/authentication/token-types) - xoxb- bot token prefix

---

### Rule 099: GitHub PAT pattern (ghp_)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects GitHub personal access tokens by their `ghp_` prefix and 36-character format.

**Pass Criteria:**
- No matches for pattern: `ghp_[a-zA-Z0-9]{36}`
- Never commit GitHub tokens to version control

**Fail Message:**
`GitHub personal access token pattern detected`

**Source:** [GitHub Token Formats](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github#githubs-token-formats) - ghp_ prefix

---

### Rule 100: Google API key pattern (AIza)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects Google API keys by their distinctive `AIza` prefix and length format.

**Pass Criteria:**
- No matches for pattern: `AIza[0-9A-Za-z\-_]{35}`
- Use environment variables for Google API keys

**Fail Message:**
`Google API key pattern detected - use environment variables`

**Source:** [Google API Key Format](https://cloud.google.com/docs/authentication/api-keys) - AIza prefix pattern

---

### Rule 101: Google OAuth client ID pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects Google OAuth client IDs by their distinctive format ending in `.apps.googleusercontent.com`.

**Pass Criteria:**
- No matches for pattern: `[0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com`
- While less sensitive than secrets, client IDs should typically be in configuration

**Fail Message:**
`Google OAuth client ID pattern detected`

**Source:** [Google OAuth2 Setup](https://developers.google.com/identity/protocols/oauth2) - .apps.googleusercontent.com pattern

---

### Rule 102: Private key detected (-----BEGIN)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects private cryptographic keys in PEM format (RSA, DSA, or generic private keys).

**Pass Criteria:**
- No matches for pattern: `-----BEGIN (RSA |DSA )?PRIVATE KEY-----`
- Never commit private keys to version control

**Fail Message:**
`Private key detected - never commit private keys`

**Source:** [RFC 7468 - PEM Encoding](https://datatracker.ietf.org/doc/html/rfc7468) - BEGIN PRIVATE KEY header

---

### Rule 103: AWS access key ID pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects AWS access key IDs by their assignment pattern and 20-character uppercase alphanumeric format.

**Pass Criteria:**
- No matches for pattern: `aws_access_key_id\s*[:=]\s*["']?[A-Z0-9]{20}["']?`
- Use environment variables or AWS IAM roles for credentials

**Fail Message:**
`AWS access key ID detected - use environment variables`

**Source:** [AWS IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html) - 20-char access key format

---

### Rule 104: AWS secret access key pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects AWS secret access keys by their assignment pattern and 40-character base64 format.

**Pass Criteria:**
- No matches for pattern: `aws_secret_access_key\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}["']?`
- Use environment variables or AWS IAM roles for credentials

**Fail Message:**
`AWS secret access key detected - use environment variables`

**Source:** [AWS IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html) - 40-char secret key format

---

## Best Practices

**Environment Variables:**
- Load all credentials from environment variables
- Use `.env` files for local development (add to `.gitignore`)
- Document required environment variables in README

**Secrets Management:**
- Use dedicated secrets management systems (HashiCorp Vault, AWS Secrets Manager, etc.)
- Rotate credentials regularly
- Use IAM roles instead of hardcoded credentials where possible

**Pre-Commit Protection:**
- Run `cclint` as a pre-commit hook to catch secrets before they're committed
- Review all warnings, even in generated code or examples
- Use `.gitignore` to exclude files containing real credentials
