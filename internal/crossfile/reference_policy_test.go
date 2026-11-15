package crossfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func writePolicyReference(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	body := "| Trigger | Target |\n|---|---|\n| review | known-skill |\n| review | missing-skill |\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func policyValidator(root string) *CrossFileValidator {
	return NewCrossFileValidator([]discovery.File{{
		Type: discovery.FileTypeSkill, RelPath: "skills/known-skill/SKILL.md",
	}}, root)
}

func assertPolicyReferences(t *testing.T, v *CrossFileValidator, root string, included bool) {
	t.Helper()
	wantIssues, wantOrphans := 0, 1
	if included {
		wantIssues, wantOrphans = 1, 0
	}
	if got := v.ValidateTriggerMaps(root); len(got) != wantIssues {
		t.Errorf("ghosts = %v; want %d", got, wantIssues)
	}
	if got := v.DetectTriggerConflicts(root); len(got) != wantIssues {
		t.Errorf("conflicts = %v; want %d", got, wantIssues)
	}
	if got := v.FindOrphanedSkills(); len(got) != wantOrphans {
		t.Errorf("orphans = %v; want %d", got, wantOrphans)
	}
}

func TestReferencePolicyExclusionsAndSnapshots(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writePolicyReference(t, filepath.Join(root, "skills/routing/references/routes.md"))
	original := policyValidator(root)
	included := original.WithReferenceSnapshot(root)
	assertPolicyReferences(t, included, root, true)

	for _, pattern := range []string{"skills/routing/**", "skills/routing/references/routes.md"} {
		t.Run(pattern, func(t *testing.T) {
			t.Parallel()
			exclude := []string{pattern}
			configured := included.WithReferencePolicy(ReferencePolicy{Exclude: exclude})
			exclude[0] = "different-path"
			assertPolicyReferences(t, configured, root, false)
			assertPolicyReferences(t, configured.WithReferenceSnapshot(root), root, false)
		})
	}
	assertPolicyReferences(t, original, root, true)
}

func TestReferencePolicySymlinks(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                                string
		ancestor, outside, follow, included bool
	}{
		{name: "file not followed"},
		{name: "ancestor not followed", ancestor: true},
		{name: "in-root file followed", follow: true, included: true},
		{name: "in-root ancestor followed", ancestor: true, follow: true, included: true},
		{name: "outside file rejected", outside: true, follow: true},
		{name: "outside ancestor rejected", ancestor: true, outside: true, follow: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			targetRoot := root
			if tc.outside {
				targetRoot = t.TempDir()
			}
			targetDir := filepath.Join(targetRoot, "reference-assets")
			writePolicyReference(t, filepath.Join(targetDir, "routes.md"))
			link := filepath.Join(root, "skills/routing/references")
			target := targetDir
			if !tc.ancestor {
				link = filepath.Join(link, "routes.md")
				target = filepath.Join(targetDir, "routes.md")
			}
			if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, link); err != nil {
				t.Fatal(err)
			}
			v := policyValidator(root).WithReferencePolicy(ReferencePolicy{FollowSymlinks: tc.follow})
			assertPolicyReferences(t, v, root, tc.included)
			assertPolicyReferences(t, v.WithReferenceSnapshot(root), root, tc.included)
		})
	}
}

func TestReferencePolicyLinkedRootKeepsLogicalPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writePolicyReference(t, filepath.Join(root, "skills/routing/references/routes.md"))
	link := filepath.Join(t.TempDir(), "linked-root")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	v := policyValidator(link)
	assertPolicyReferences(t, v, link, true)
	got := v.ValidateTriggerMaps(link)
	if len(got) != 1 || got[0].File != "skills/routing/references/routes.md" {
		t.Fatalf("logical diagnostic paths lost: %v", got)
	}
}
