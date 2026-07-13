package common

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Paging struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// PagingFromQuery đọc ?page=&page_size= (mặc định 1/50, trần 100).
func PagingFromQuery(c *gin.Context) Paging {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if size < 1 {
		size = 50
	}
	if size > 100 {
		size = 100
	}
	return Paging{Page: page, PageSize: size}
}

func (p Paging) Offset() int { return (p.Page - 1) * p.PageSize }
