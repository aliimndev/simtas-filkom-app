// Package grading implements the fixed-weight score calculations used by the
// seminar (Job 08) and defense (Job 09) modules.
package grading

import (
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// ComponentScore is a single examiner's score for one grading component.
type ComponentScore struct {
	Name   string
	Weight float64
	Score  float64
}

// CalculateExaminerScore returns the weighted score of one examiner:
// Σ (score_component × weight / 100). The result is clamped to [0, 100].
func CalculateExaminerScore(components []ComponentScore) float64 {
	var total float64
	for _, c := range components {
		total += c.Score * c.Weight / 100.0
	}
	return clamp(total)
}

// CalculateFinalScore averages the weighted score of every examiner:
// Nilai_Akhir = Σ Nilai_Penguji_i / jumlah_penguji.
func CalculateFinalScore(examinerScores []float64) float64 {
	if len(examinerScores) == 0 {
		return 0
	}
	var sum float64
	for _, s := range examinerScores {
		sum += s
	}
	return clamp(sum / float64(len(examinerScores)))
}

// CalculateFinalFromScores groups persisted score rows by examiner, computes
// each examiner's weighted score, and averages them. Useful when scores are
// stored flat (seminar_scores / defense_scores).
func CalculateFinalFromScores(weights []entity.GradingComponent, scores []WeightedScoreRow) float64 {
	byExaminer := map[string][]ComponentScore{}
	order := []string{}
	for _, row := range scores {
		if _, ok := byExaminer[row.ExaminerID]; !ok {
			order = append(order, row.ExaminerID)
		}
		byExaminer[row.ExaminerID] = append(byExaminer[row.ExaminerID], ComponentScore{
			Name:   row.ComponentName,
			Weight: weightFor(weights, row.ComponentName, row.ComponentWeight),
			Score:  row.Score,
		})
	}

	examinerScores := make([]float64, 0, len(order))
	for _, id := range order {
		examinerScores = append(examinerScores, CalculateExaminerScore(byExaminer[id]))
	}
	return CalculateFinalScore(examinerScores)
}

// WeightedScoreRow is a flat score row used by CalculateFinalFromScores.
type WeightedScoreRow struct {
	ExaminerID      string
	ComponentName   string
	ComponentWeight float64
	Score           float64
}

func weightFor(weights []entity.GradingComponent, name string, fallback float64) float64 {
	for _, w := range weights {
		if w.Name == name {
			return w.Weight
		}
	}
	return fallback
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
