package crossfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestReferenceSnapshotEquivalenceAndLifecycle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	otherRoot := t.TempDir()
	path := filepath.Join(root, "skills", "routing", "references", "routes.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("| Trigger | Target |\n|---|---|\n| review | known-skill |\n| review | missing-skill |\n")
	files := []discovery.File{{Type: discovery.FileTypeSkill, RelPath: "skills/known-skill/SKILL.md"}}
	original := NewCrossFileValidator(files, root)
	batch := original.WithReferenceSnapshot(root)

	if got, want := batch.FindOrphanedSkills(), original.FindOrphanedSkills(); !reflect.DeepEqual(got, want) {
		t.Fatalf("orphans differ: got %v, want %v", got, want)
	}
	if got, want := batch.ValidateTriggerMaps(root), original.ValidateTriggerMaps(root); !reflect.DeepEqual(got, want) {
		t.Fatalf("ghosts differ: got %v, want %v", got, want)
	}
	if got, want := batch.DetectTriggerConflicts(root), original.DetectTriggerConflicts(root); !reflect.DeepEqual(got, want) {
		t.Fatalf("conflicts differ: got %v, want %v", got, want)
	}

	write("| Trigger | Target |\n|---|---|\n| review | known-skill |\n")
	if got := original.ValidateTriggerMaps(root); len(got) != 0 {
		t.Fatalf("standalone validator retained stale ghosts: %v", got)
	}
	if got := original.DetectTriggerConflicts(root); len(got) != 0 {
		t.Fatalf("standalone validator retained stale conflicts: %v", got)
	}
	if got := batch.ValidateTriggerMaps(root); len(got) != 1 {
		t.Fatalf("batch snapshot lost captured ghosts: %v", got)
	}
	if got := batch.DetectTriggerConflicts(root); len(got) != 1 {
		t.Fatalf("batch snapshot lost captured conflicts: %v", got)
	}
	if got := original.WithReferenceSnapshot(root).ValidateTriggerMaps(root); len(got) != 0 {
		t.Fatalf("next batch retained stale ghosts: %v", got)
	}
	if got := batch.ValidateTriggerMaps(otherRoot); len(got) != 0 {
		t.Fatalf("snapshot leaked into other root: %v", got)
	}
}
