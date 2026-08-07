package statemachine

import "testing"

func TestTitleChangeRequestCanTransition(t *testing.T) {
	tests := []struct {
		from string
		to   string
		want bool
	}{
		{"PENDING", "APPROVED", true},
		{"PENDING", "REJECTED", true},
		{"PENDING", "CANCELLED", true},
		{"PENDING", "PENDING", false},
		{"APPROVED", "PENDING", false},
		{"APPROVED", "REJECTED", false},
		{"APPROVED", "CANCELLED", false},
		{"REJECTED", "APPROVED", false},
		{"REJECTED", "CANCELLED", false},
		{"CANCELLED", "PENDING", false},
		{"CANCELLED", "APPROVED", false},
	}

	for _, tt := range tests {
		if got := TitleChangeRequestCanTransition(tt.from, tt.to); got != tt.want {
			t.Errorf("TitleChangeRequestCanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestTitleChangeRequestValidStatus(t *testing.T) {
	valid := []string{"PENDING", "APPROVED", "REJECTED", "CANCELLED"}
	for _, s := range valid {
		if !TitleChangeRequestValidStatus(s) {
			t.Errorf("TitleChangeRequestValidStatus(%q) should be true", s)
		}
	}
	if TitleChangeRequestValidStatus("nonsense") {
		t.Error("TitleChangeRequestValidStatus(nonsense) should be false")
	}
}
