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
