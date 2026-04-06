package dto

type SearchParams struct {
	Query   string `json:"q"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	SortBy  string `json:"sort_by"` // relevance, recent, views
}

func (p *SearchParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 50 {
		p.PerPage = 20
	}
	switch p.SortBy {
	case "relevance", "recent", "views":
	default:
		p.SortBy = "relevance"
	}
}

type PaginationParams struct {
	Page    int
	PerPage int
}

func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 50 {
		p.PerPage = 20
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}
