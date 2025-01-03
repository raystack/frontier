package pagination

const (
	DefaultPageSize = 1000
	DefaultPageNum  = 1
)

type Pagination struct {
	PageNum  int32
	PageSize int32
	Count    int32 // total number of records in DB
}

func NewPagination(pageNum, pageSize int32) *Pagination {
	if pageNum == 0 {
		pageNum = DefaultPageNum
	}
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}

	return &Pagination{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

func (p *Pagination) Offset() int32 {
	return p.PageSize * (p.PageNum - 1)
}

func (p *Pagination) SetCount(count int32) {
	p.Count = count
}
