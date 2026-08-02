package entity

// GradingComponent describes one scored component with its fixed weight (v1.0).
type GradingComponent struct {
	Name   string
	Weight float64
}

// SeminarGradingComponents are the fixed weights used for seminar proposal
// scoring (Job 08). Weights sum to 100.
var SeminarGradingComponents = []GradingComponent{
	{Name: "Presentasi", Weight: 30.0},
	{Name: "Penguasaan Materi", Weight: 30.0},
	{Name: "Kualitas Naskah", Weight: 25.0},
	{Name: "Kemampuan Menjawab", Weight: 15.0},
}

// DefenseGradingComponents are the fixed weights used for thesis defense
// scoring (Job 09). They match the seminar weights in v1.0.
var DefenseGradingComponents = []GradingComponent{
	{Name: "Presentasi", Weight: 30.0},
	{Name: "Penguasaan Materi", Weight: 30.0},
	{Name: "Kualitas Naskah", Weight: 25.0},
	{Name: "Kemampuan Menjawab", Weight: 15.0},
}

// ValidGradingComponents validates a component name against the given list.
func ValidGradingComponents(components []GradingComponent, name string) bool {
	for _, c := range components {
		if c.Name == name {
			return true
		}
	}
	return false
}
