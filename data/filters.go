package data

import (
	"math"
	"strings"

	"github.com/eze8789/movies-api/validator"
)

type Filters struct {
	Page     int
	PageSize int
	Sort     string
	SortSafe []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func ValidateFilter(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "page needs to be greater than 0")
	v.Check(f.Page <= 1000, "page", "page needs to be lower than 1000") //nolint:gomnd

	v.Check(f.PageSize >= 1, "page_size", "page_size can not be less than 1")
	v.Check(f.PageSize <= 100, "page_size", "page_size can not be greater than 100") //nolint:gomnd

	v.Check(v.In(f.Sort, f.SortSafe...), "sort", "invalid sort value")
}

//nolint:gocritic
func (f Filters) sortColumn() string {
	// s := strings.Split(f.Sort, "-")
	// if len(s) == 1 {
	// 	return s[0]
	// }
	// return s[1]

	s := strings.Split(f.Sort, "-")
	if len(s) == 1 {
		return s[0]
	}
	return s[1]
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calcMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
