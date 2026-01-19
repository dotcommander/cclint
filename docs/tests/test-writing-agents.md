# Writing Agents Guide Tests

Tests for `docs/guides/writing-agents.md` acceptance criteria.

---

## Test: Required Sections Explained

**Acceptance Criterion:** Given docs/guides/writing-agents.md exists, when read, then required sections (Foundation, Workflow) are explained.

### Test Procedure

1. Read `docs/guides/writing-agents.md`
2. Search for "Foundation Section" heading
3. Verify explanation of Foundation section exists
4. Search for "Workflow Section" heading
5. Verify explanation of Workflow section exists

### Expected Result

- File contains section explaining Foundation section purpose and required elements
- File contains section explaining Workflow section with phase-based pattern
- Both sections include examples

---

## Test: 200-Line Limit Rationale

**Acceptance Criterion:** Given writing-agents.md, when scanning content, then 200-line limit rationale is documented.

### Test Procedure

1. Read `docs/guides/writing-agents.md`
2. Search for "Why 200 Lines?" heading or similar
3. Verify explanation of thin-agent/fat-skill pattern
4. Verify tolerance documentation (220 lines)

### Expected Result

- File explains the 200-line limit enforces thin-agent/fat-skill pattern
- Rationale includes: agents orchestrate, skills contain methodology
- Tolerance thresholds documented (excellent/good/OK/over limit)
- Table showing size tiers with line counts

---

## Test: Complete Passing Agent Example

**Acceptance Criterion:** Given writing-agents.md, when checking examples, then complete passing agent example is included.

### Test Procedure

1. Read `docs/guides/writing-agents.md`
2. Search for "Complete Example" heading or similar
3. Verify example includes frontmatter with all required fields
4. Verify example includes Foundation section
5. Verify example includes Workflow section
6. Verify example follows all best practices documented

### Expected Result

- Complete agent example with valid YAML frontmatter
- Contains: name, description (with PROACTIVELY), model, tools
- Contains ## Foundation section with purpose/capabilities/constraints
- Contains ## Workflow section with numbered phases
- Contains ## Success Criteria and ## Edge Cases
- Example is under 200 lines
- Example demonstrates all DO/DON'T principles from guide

---

## Test: Validation Checklist

**Bonus:** Verify the guide includes a validation checklist for authors.

### Test Procedure

1. Read `docs/guides/writing-agents.md`
2. Search for "Validation Checklist" heading
3. Verify checklist covers all major rules

### Expected Result

- Checklist with 10+ items covering name, description, size, sections
- Checkbox format for easy validation
- Covers both required fields and anti-patterns
