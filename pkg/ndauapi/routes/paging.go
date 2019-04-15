package routes

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

// paging queries
// limit -- max number of items to return
// after -- string ID after which to start the query
// For queries that have metadata, the metadata includes a "next" that returns the query parameters
// to page the data; if next is empty, then there are no more results.
// we name the return values for documentation purposes
func getPagingParams(r *http.Request, maxlimit int) (limit int, after string, err error) {
	err = nil

	// Paging is optional. Default to returning the max.
	limit = maxlimit
	after = ""

	slimit := r.URL.Query().Get("limit")
	if slimit != "" {
		n, err := strconv.Atoi(slimit)
		if err != nil {
			return limit, after, errors.Wrap(err, "limit must be a number")
		}
		limit = n
	}

	after = r.URL.Query().Get("after")

	if limit <= 0 {
		limit = maxlimit
	}

	return limit, after, nil
}

// Returns non-empty error message if the given request was bad.
func getOldPagingParams(r *http.Request) (pageIndex int, pageSize int, errMsg string, err error) {
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

		// Don't let the user set a page size larger than the max we allow.
		if pageSize > MaximumRange {
			pageSize = MaximumRange
		}
	}

	// Underlying search implementation handles pageSize == 0 to mean "get all results", but we
	// don't know how large that list could be.  Instead, when page size isn't specified, use the
	// max allowed and start at page 0 as a special behavior (documented in the README.md file).
	if pageSize == 0 {
		pageIndex = 0
		pageSize = MaximumRange
	}

	return
}
