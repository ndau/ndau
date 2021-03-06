package sdk

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"strings"

	"github.com/ndau/json2msgp"
	"github.com/ndau/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// Sysvars gets the list of requested system variables, marshaled
func (c *Client) Sysvars(vars ...string) (svs map[string][]byte, err error) {
	svsj := make(map[string]interface{})
	var url string
	if len(vars) == 0 {
		url = c.URL("system/all")
	} else {
		url = c.URL("system/get/%s", strings.Join(vars, ","))
	}
	err = c.get(&svsj, url)
	if err != nil {
		err = errors.Wrap(err, "getting sysvars")
		return
	}
	svs = make(map[string][]byte)
	for name, jsdata := range svsj {
		var data []byte
		data, err = json2msgp.Convert(jsdata, nil)
		if err != nil {
			err = fmt.Errorf("failed to convert sysvar %s to msgpack", name)
			return
		}
		svs[name] = data
	}
	return
}

// Sysvars gets the list of requested system variables, marshaled
func Sysvars(node *Client, vars ...string) (map[string][]byte, error) {
	return node.Sysvars(vars...)
}

// Sysvar gets a single system variable given its name and an example of its type
//
// The example is populated with the appropriate data
func (c *Client) Sysvar(name string, example msgp.Unmarshaler) error {
	svs, err := c.Sysvars(name)
	if err != nil {
		return errors.Wrap(err, "getting system variable")
	}
	sv, ok := svs[name]
	if !ok {
		return errors.New("API did not return requested system variable")
	}
	_, err = example.UnmarshalMsg(sv)
	return errors.Wrap(err, "unmarshaling")
}

// Sysvar gets a single system variable given its name and an example of its type
//
// The example is populated with the appropriate data
func Sysvar(node *Client, name string, example msgp.Unmarshaler) error {
	return node.Sysvar(name, example)
}

// SysvarHistory gets the value history of the given sysvar.
//
// Pass in 0,0 for the paging params to get the entire history.
func (c *Client) SysvarHistory(name string, after uint64, limit int) (resp *query.SysvarHistoryResponse, err error) {
	resp = new(query.SysvarHistoryResponse)
	err = c.get(resp, c.URLP(
		params{"after": after, "limit": limit},
		"system/history/%s", name,
	))
	err = errors.Wrap(err, "getting sysvar history")
	return
}

// SysvarHistory gets the value history of the given sysvar.
//
// Pass in 0,0 for the paging params to get the entire history.
func SysvarHistory(node *Client, name string, after uint64, limit int) (*query.SysvarHistoryResponse, error) {
	return node.SysvarHistory(name, after, limit)
}
