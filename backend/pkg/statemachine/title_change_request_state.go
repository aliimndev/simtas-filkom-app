package statemachine

// ValidTransitions defines allowed title change request status transitions.
//
//	PENDING   -> APPROVED / REJECTED / CANCELLED
//	APPROVED  -> (none)
//	REJECTED  -> (none)
//	CANCELLED -> (none)
var TitleChangeRequestValidTransitions = map[string][]string{
	"PENDING":   {"APPROVED", "REJECTED", "CANCELLED"},
	"APPROVED":  {},
	"REJECTED":  {},
	"CANCELLED": {},
}

// TitleChangeRequestCanTransition reports whether a title change request may
// move from status `from` to `to`.
func TitleChangeRequestCanTransition(from, to string) bool {
	for _, next := range TitleChangeRequestValidTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// TitleChangeRequestValidStatus reports whether a status is a known title change
// request status.
func TitleChangeRequestValidStatus(status string) bool {
	_, ok := TitleChangeRequestValidTransitions[status]
	return ok
}
