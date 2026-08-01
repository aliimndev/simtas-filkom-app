package statemachine

import "testing"

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from string
		to   string
		want bool
	}{
		{"submitted", "approved", true},
		{"submitted", "rejected", true},
		{"submitted", "in_progress", false},
		{"approved", "in_progress", true},
		{"approved", "approved", false},
		{"in_progress", "seminar_ready", true},
		{"in_progress", "submitted", false},
		{"seminar_ready", "seminar_done", true},
		{"seminar_done", "defense_ready", true},
		{"defense_ready", "defense_done", true},
		{"defense_done", "graduated", true},
		{"submitted", "graduated", false},
		{"graduated", "graduated", false},
		{"approved", "rejected", false},
		{"rejected", "approved", false},
	}

	for _, tt := range tests {
		if got := CanTransition(tt.from, tt.to); got != tt.want {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestCanTransitionCancelledFromAnyStatus(t *testing.T) {
	statuses := []string{
		"submitted", "approved", "rejected", "in_progress",
		"seminar_ready", "seminar_done", "defense_ready", "defense_done", "graduated",
	}
	for _, s := range statuses {
		if s == "cancelled" {
			continue
		}
		if !CanTransition(s, "cancelled") {
			t.Errorf("CanTransition(%q, cancelled) = false, want true", s)
		}
	}
	if CanTransition("cancelled", "cancelled") {
		t.Error("CanTransition(cancelled, cancelled) should be false")
	}
}

func TestValidStatus(t *testing.T) {
	if !ValidStatus("submitted") || !ValidStatus("graduated") {
		t.Error("expected submitted/graduated to be valid statuses")
	}
	if ValidStatus("nonsense") {
		t.Error("expected nonsense to be invalid")
	}
}
