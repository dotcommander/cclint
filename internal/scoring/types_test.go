package scoring

import (
	"testing"
)

func TestTierFromScore(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		wantTier string
	}{
		{"A tier - exact boundary", 85, "A"},
		{"A tier - high score", 100, "A"},
		{"A tier - above boundary", 90, "A"},
		{"B tier - exact boundary", 70, "B"},
		{"B tier - mid range", 75, "B"},
		{"B tier - upper range", 84, "B"},
		{"C tier - exact boundary", 50, "C"},
		{"C tier - mid range", 60, "C"},
		{"C tier - upper range", 69, "C"},
		{"D tier - exact boundary", 30, "D"},
		{"D tier - mid range", 40, "D"},
		{"D tier - upper range", 49, "D"},
		{"F tier - boundary", 29, "F"},
		{"F tier - low score", 10, "F"},
		{"F tier - zero", 0, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TierFromScore(tt.score)
			if got != tt.wantTier {
				t.Errorf("TierFromScore(%d) = %q, want %q", tt.score, got, tt.wantTier)
			}
		})
	}
}

func TestNewQualityScore(t *testing.T) {
	tests := []struct {
		name          string
		structural    int
		practices     int
		composition   int
		documentation int
		wantOverall   int
		wantTier      string
	}{
		{
			name:          "Perfect score",
			structural:    40,
			practices:     40,
			composition:   10,
			documentation: 10,
			wantOverall:   100,
			wantTier:      "A",
		},
		{
			name:          "A tier",
			structural:    35,
			practices:     35,
			composition:   8,
			documentation: 8,
			wantOverall:   86,
			wantTier:      "A",
		},
		{
			name:          "B tier",
			structural:    30,
			practices:     25,
			composition:   8,
			documentation: 7,
			wantOverall:   70,
			wantTier:      "B",
		},
		{
			name:          "C tier",
			structural:    20,
			practices:     20,
			composition:   6,
			documentation: 4,
			wantOverall:   50,
			wantTier:      "C",
		},
		{
			name:          "D tier",
			structural:    15,
			practices:     10,
			composition:   3,
			documentation: 2,
			wantOverall:   30,
			wantTier:      "D",
		},
		{
			name:          "F tier",
			structural:    10,
			practices:     5,
			composition:   0,
			documentation: 0,
			wantOverall:   15,
			wantTier:      "F",
		},
		{
			name:          "Zero score",
			structural:    0,
			practices:     0,
			composition:   0,
			documentation: 0,
			wantOverall:   0,
			wantTier:      "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := []ScoringMetric{
				{Category: "structural", Name: "test", Points: tt.structural, MaxPoints: 40, Passed: true},
			}
			score := NewQualityScore(tt.structural, tt.practices, tt.composition, tt.documentation, details)

			if score.Overall != tt.wantOverall {
				t.Errorf("Overall = %d, want %d", score.Overall, tt.wantOverall)
			}
			if score.Tier != tt.wantTier {
				t.Errorf("Tier = %q, want %q", score.Tier, tt.wantTier)
			}
			if score.Structural != tt.structural {
				t.Errorf("Structural = %d, want %d", score.Structural, tt.structural)
			}
			if score.Practices != tt.practices {
				t.Errorf("Practices = %d, want %d", score.Practices, tt.practices)
			}
			if score.Composition != tt.composition {
				t.Errorf("Composition = %d, want %d", score.Composition, tt.composition)
			}
			if score.Documentation != tt.documentation {
				t.Errorf("Documentation = %d, want %d", score.Documentation, tt.documentation)
			}
			if len(score.Details) != len(details) {
				t.Errorf("Details count = %d, want %d", len(score.Details), len(details))
			}
		})
	}
}

func TestScoringMetric(t *testing.T) {
	metric := ScoringMetric{
		Category:  "structural",
		Name:      "Has name",
		Points:    10,
		MaxPoints: 10,
		Passed:    true,
		Note:      "Present",
	}

	if metric.Category != "structural" {
		t.Errorf("Category = %q, want %q", metric.Category, "structural")
	}
	if metric.Name != "Has name" {
		t.Errorf("Name = %q, want %q", metric.Name, "Has name")
	}
	if metric.Points != 10 {
		t.Errorf("Points = %d, want %d", metric.Points, 10)
	}
	if metric.MaxPoints != 10 {
		t.Errorf("MaxPoints = %d, want %d", metric.MaxPoints, 10)
	}
	if !metric.Passed {
		t.Error("Passed = false, want true")
	}
	if metric.Note != "Present" {
		t.Errorf("Note = %q, want %q", metric.Note, "Present")
	}
}
