package pagination

import (
	"math"
)

const (
	DefaultPageSize = 50
	DefaultPageNum  = 1
)

type Pagination struct {
	PageNum    uint
	PageSize   uint
	TotalPages uint
}

func NewPagination(pageNum, pageSize uint) *Pagination {
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

func (p *Pagination) Offset() uint {
	return p.PageSize * (p.PageNum - 1)
}

func (p *Pagination) SetTotalPages(totalCount uint) {
	p.TotalPages = uint(math.Ceil(float64(totalCount) / float64(p.PageSize)))
}
