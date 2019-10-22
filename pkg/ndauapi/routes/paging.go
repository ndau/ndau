package routes

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

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
	qp := getQueryParms(r)

	slimit := qp["limit"]
	if slimit != "" {
		n, err := strconv.Atoi(slimit)
		if err != nil {
			return limit, after, errors.Wrap(err, "limit must be a number")
		}
		limit = n
	}

	after = qp["after"]

	if limit <= 0 {
		limit = maxlimit
	}

	return limit, after, nil
}
