package routes

import (
	"net/http"
	"strconv"
)

// Returns non-empty error message if the given request was bad.
func getPagingParams(r *http.Request) (pageIndex int, pageSize int, errMsg string, err error) {
	errMsg = ""
	err = nil

	// Paging is optional.  Default to returning all history in a single page.
	pageIndex = 0
	pageSize = 0

	pageIndexString := r.URL.Query().Get("pageindex")
	if pageIndexString != "" {
		var pageIndex64 int64
		pageIndex64, err = strconv.ParseInt(pageIndexString, 10, 32)
		if err != nil {
			errMsg = "pageindex must be a valid number"
			return
		}
		pageIndex = int(pageIndex64)
	}

	pageSizeString := r.URL.Query().Get("pagesize")
	if pageSizeString != "" {
		var pageSize64 int64
		pageSize64, err = strconv.ParseInt(pageSizeString, 10, 32)
		if err != nil {
			errMsg = "pagesize must be a valid number"
			return
		}
		if pageSize64 < 0 {
			errMsg = "pagesize must be non-negative"
			return
		}
		pageSize = int(pageSize64)
	}

	return
}
