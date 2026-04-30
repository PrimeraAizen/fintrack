package domain

type PageParams struct {
	Page    int
	PerPage int
}

func (p PageParams) Normalize() PageParams {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	return p
}

func (p PageParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}
