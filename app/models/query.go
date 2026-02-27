package models

type QueryParams struct {
	Page  int64
	Limit int64
}

func (q QueryParams) Normalize() QueryParams {
	out := q
	if out.Page <= 0 {
		out.Page = 1
	}
	if out.Limit <= 0 || out.Limit > 100 {
		out.Limit = 20
	}
	return out
}

func (q QueryParams) Skip() int64 {
	n := q.Normalize()
	return (n.Page - 1) * n.Limit
}
