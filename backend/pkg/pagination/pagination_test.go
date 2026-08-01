package pagination

import "testing"

func TestNewPagination(t *testing.T) {
	tests := []struct {
		name          string
		page, perPage int
		total         int64
		wantPage      int
		wantPerPage   int
		wantTotalPg   int
	}{
		{"normal", 1, 20, 85, 1, 20, 5},
		{"exact multiple", 2, 10, 20, 2, 10, 2},
		{"zero total", 1, 20, 0, 1, 20, 0},
		{"page below 1", 0, 20, 50, 1, 20, 3},
		{"perPage below 1", 1, 0, 30, 1, 20, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPagination(tt.page, tt.perPage, tt.total)
			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.PerPage != tt.wantPerPage {
				t.Errorf("PerPage = %d, want %d", p.PerPage, tt.wantPerPage)
			}
			if p.TotalPages != tt.wantTotalPg {
				t.Errorf("TotalPages = %d, want %d", p.TotalPages, tt.wantTotalPg)
			}
		})
	}
}

func TestPaginationOffset(t *testing.T) {
	p := NewPagination(3, 20, 100)
	if got := p.Offset(); got != 40 {
		t.Errorf("Offset() = %d, want 40", got)
	}
}
