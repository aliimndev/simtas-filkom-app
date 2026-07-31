package dto

type PaginationRequest struct {
	Page    int `form:"page" json:"page" validate:"min=1"`
	PerPage int `form:"per_page" json:"per_page" validate:"min=1,max=100"`
}

type IDParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}
