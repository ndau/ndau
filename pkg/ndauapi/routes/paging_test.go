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
	"net/http/httptest"
	"testing"
)

func Test_getPagingParams(t *testing.T) {
	tests := []struct {
		target    string
		wantLimit int
		wantAfter string
		wantErr   bool
	}{
		{"/", MaximumRange, "", false},
		{"/?limit=3", 3, "", false},
		{"/?after=asdf", MaximumRange, "asdf", false},
		{"/?limit=3&after=asdf", 3, "asdf", false},
		{"/?lImIt=3", 3, "", false},
		{"/?aFtEr=asdf", MaximumRange, "asdf", false},
		{"/?lImIt=3&aFtEr=asdf", 3, "asdf", false},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			gotLimit, gotAfter, err := getPagingParams(httptest.NewRequest("", tt.target, nil), MaximumRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPagingParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLimit != tt.wantLimit {
				t.Errorf("getPagingParams() gotLimit = %v, want %v", gotLimit, tt.wantLimit)
			}
			if gotAfter != tt.wantAfter {
				t.Errorf("getPagingParams() gotAfter = %v, want %v", gotAfter, tt.wantAfter)
			}
		})
	}
}
