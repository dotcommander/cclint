# Writing Skills Guide Tests

Tests for `docs/guides/writing-skills.md` acceptance criteria.

---

## Test: Quick Reference Table Explained

**Acceptance Criterion:** Given docs/guides/writing-skills.md exists, when read, then Quick Reference table requirement is explained.

### Test Procedure

1. Read `docs/guides/writing-skills.md`
2. Search for "Quick Reference Table" heading
3. Verify explanation of Quick Reference section exists
4. Verify documentation of table structure and purpose

### Expected Result

- File contains section explaining Quick Reference table requirement
- Explains semantic routing purpose
- Shows required table structure with Intent/Action columns
- Includes best practices for writing table entries

---

## Test: 500-Line Limit Rationale

**Acceptance Criterion:** Given writing-skills.md, when scanning content, then 500-line limit rationale is documented.

### Test Procedure

1. Read `docs/guides/writing-skills.md`
2. Search for "Why 500 Lines?" heading or similar
3. Verify explanation of thin skill → fat references pattern
4. Verify tolerance documentation (550 lines)

### Expected Result

- File explains the 500-line limit enforces progressive disclosure
- Rationale includes: SKILL.md as entry point, references/ for heavy content
- Tolerance thresholds documented (excellent/good/OK/over limit)
- Table showing size tiers with line counts
- Explains when to move content to references/ subdirectory

---

## Test: Complete Passing Skill Example

**Acceptance Criterion:** Given writing-skills.md, when checking examples, then complete passing skill example is included.

### Test Procedure

1. Read `docs/guides/writing-skills.md`
2. Search for "Complete Example" heading or similar
3. Verify example includes frontmatter with all required fields
4. Verify example includes Quick Reference table
5. Verify example includes When to Use section
6. Verify example follows all best practices documented

### Expected Result

- Complete skill example with valid YAML frontmatter
- Contains: name, description (with trigger phrases)
- Contains ## Quick Reference table with Intent/Action columns
- Contains ## When to Use section with bullet points
- Contains ## Anti-Patterns section with table
- Contains ## Examples section with concrete examples
- Example is under 500 lines (documented line count)
- Example demonstrates all DO/DON'T principles from guide
- Example shows references/ usage pattern
