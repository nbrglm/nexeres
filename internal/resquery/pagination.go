package resquery

type QueryPagination struct {
	// Page is the page number to retrieve (0-based).
	Page     int `json:"page" binding:"required,min=0"`
	PageSize int `json:"pageSize" binding:"required,min=1,max=100"`
}

func NewQueryPagination(page, pageSize int) *QueryPagination {
	return &QueryPagination{
		Page:     page,
		PageSize: pageSize,
	}
}

func DefaultQueryPagination() *QueryPagination {
	return &QueryPagination{
		Page:     0,
		PageSize: 10,
	}
}
