// Package statemachine provides reusable workflow state machines.
package statemachine

// ValidTransitions defines the allowed thesis status transitions (Job 05).
//
//	submitted     → approved / rejected / cancelled
//	approved      → in_progress / cancelled
//	in_progress   → seminar_ready / cancelled
//	seminar_ready → seminar_done
//	seminar_done  → defense_ready
//	defense_ready → defense_done
//	defense_done  → graduated
//
// cancelled is reachable from any status (admin/kaprodi), which is handled
// separately in CanTransition so the map stays declarative.
var ValidTransitions = map[string][]string{
	"submitted":     {"approved", "rejected", "cancelled"},
	"approved":      {"in_progress", "cancelled"},
	"in_progress":   {"seminar_ready", "cancelled"},
	"seminar_ready": {"seminar_done"},
	"seminar_done":  {"defense_ready"},
	"defense_ready": {"defense_done"},
	"defense_done":  {"graduated"},
	// Terminal / sink states: they appear as keys so ValidStatus recognizes them,
	// but they have no outgoing transitions (new submission starts from scratch).
	"rejected":  {},
	"graduated": {},
	"cancelled": {},
}

// CanTransition reports whether a thesis may move from status `from` to `to`.
func CanTransition(from, to string) bool {
	// Cancelled is allowed from any status except itself.
	if to == "cancelled" {
		return from != "cancelled"
	}
	for _, next := range ValidTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// ValidStatus reports whether a status is a known thesis status.
func ValidStatus(status string) bool {
	_, ok := ValidTransitions[status]
	return ok
}
