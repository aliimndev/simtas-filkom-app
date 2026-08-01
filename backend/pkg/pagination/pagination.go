package pagination

// Pagination is a reusable pagination metadata container.
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPagination builds Pagination from raw page/perPage/total values.
// Guards against invalid inputs (page < 1, perPage < 1).
func NewPagination(page, perPage int, total int64) Pagination {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}

	return Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Offset returns the SQL OFFSET for the current page.
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}
