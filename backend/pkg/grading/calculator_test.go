package grading

import (
	"math"
	"testing"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

const epsilon = 1e-9

func assertAlmostEqual(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > epsilon {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestCalculateExaminerScore checks the weighted Σ(score×weight/100) formula.
func TestCalculateExaminerScore(t *testing.T) {
	got := CalculateExaminerScore([]ComponentScore{
		{Name: "Presentasi", Weight: 30, Score: 80},
		{Name: "Pembahasan", Weight: 40, Score: 75},
		{Name: "Penguasaan", Weight: 30, Score: 90},
	})
	// 80*0.30 + 75*0.40 + 90*0.30 = 24 + 30 + 27 = 81
	assertAlmostEqual(t, got, 81.0)
}

func TestCalculateExaminerScore_ClampedTo100(t *testing.T) {
	// Weights sum to 100, scores all 100 → exactly 100.
	got := CalculateExaminerScore([]ComponentScore{
		{Name: "A", Weight: 50, Score: 100},
		{Name: "B", Weight: 50, Score: 100},
	})
	assertAlmostEqual(t, got, 100.0)
}

func TestCalculateExaminerScore_EmptyClampedTo0(t *testing.T) {
	assertAlmostEqual(t, CalculateExaminerScore(nil), 0.0)
}

// TestCalculateFinalScore_TwoExaminers — Job 23: 2 penguji, nilai berbeda.
func TestCalculateFinalScore_TwoExaminers(t *testing.T) {
	// Examiner 1 = 90, Examiner 2 = 70 → average 80.
	assertAlmostEqual(t, CalculateFinalScore([]float64{90, 70}), 80.0)
	// Examiner 1 = 81, Examiner 2 = 75 → average 78.
	assertAlmostEqual(t, CalculateFinalScore([]float64{81, 75}), 78.0)
}

// TestCalculateFinalScore_Boundaries — nilai akhir tepat di batas kelulusan
// (threshold dipakai oleh GetGradeCategory di defense/seminar usecase).
func TestCalculateFinalScore_Boundaries(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"below pass", 59.9},
		{"exactly pass", 60.0},
		{"below revision threshold", 74.9},
		{"exactly revision threshold", 75.0},
		{"perfect", 100.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateFinalScore([]float64{tt.score})
			assertAlmostEqual(t, got, tt.score)
		})
	}
}

// TestCalculateFinalScore_SingleExaminer — edge case minimum penguji.
func TestCalculateFinalScore_SingleExaminer(t *testing.T) {
	assertAlmostEqual(t, CalculateFinalScore([]float64{85}), 85.0)
}

// TestCalculateFinalScore_NoExaminers — empty input must not divide by zero.
func TestCalculateFinalScore_NoExaminers(t *testing.T) {
	assertAlmostEqual(t, CalculateFinalScore(nil), 0.0)
}

// TestCalculateFinalScore_ClampedTo100 — nilai tidak melebihi 100.
func TestCalculateFinalScore_ClampedTo100(t *testing.T) {
	assertAlmostEqual(t, CalculateFinalScore([]float64{110, 120}), 100.0)
}

// TestCalculateFinalScore_ClampedTo0 — nilai tidak kurang dari 0.
func TestCalculateFinalScore_ClampedTo0(t *testing.T) {
	assertAlmostEqual(t, CalculateFinalScore([]float64{-10, -20}), 0.0)
}

// TestCalculateFinalFromScores groups flat score rows by examiner and applies
// the configured component weights.
func TestCalculateFinalFromScores(t *testing.T) {
	weights := []entity.GradingComponent{
		{Name: "Presentasi", Weight: 30},
		{Name: "Pembahasan", Weight: 40},
		{Name: "Penguasaan", Weight: 30},
	}
	rows := []WeightedScoreRow{
		{ExaminerID: "penguji-1", ComponentName: "Presentasi", Score: 80},
		{ExaminerID: "penguji-1", ComponentName: "Pembahasan", Score: 75},
		{ExaminerID: "penguji-1", ComponentName: "Penguasaan", Score: 90},
		{ExaminerID: "penguji-2", ComponentName: "Presentasi", Score: 70},
		{ExaminerID: "penguji-2", ComponentName: "Pembahasan", Score: 80},
		{ExaminerID: "penguji-2", ComponentName: "Penguasaan", Score: 60},
	}
	// Penguji 1: 24+30+27 = 81; Penguji 2: 21+32+18 = 71; average = 76.
	assertAlmostEqual(t, CalculateFinalFromScores(weights, rows), 76.0)
}

// TestCalculateFinalFromScores_FallbackWeight — rows with an unknown component
// fall back to their own ComponentWeight.
func TestCalculateFinalFromScores_FallbackWeight(t *testing.T) {
	rows := []WeightedScoreRow{
		{ExaminerID: "e1", ComponentName: "unknown", ComponentWeight: 100, Score: 80},
	}
	assertAlmostEqual(t, CalculateFinalFromScores(nil, rows), 80.0)
}

// TestCalculateFinalFromScores_Empty — no rows → 0, no panic.
func TestCalculateFinalFromScores_Empty(t *testing.T) {
	assertAlmostEqual(t, CalculateFinalFromScores(nil, nil), 0.0)
}
