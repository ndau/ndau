package tool

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/oneiro-ndev/tendermint.0.32.3/rpc/client"
	rpctypes "github.com/oneiro-ndev/tendermint.0.32.3/rpc/core/types"
	"github.com/tinylib/msgp/msgp"
)

// Sysvars gets the version the connected node is running
func Sysvars(node client.ABCIClient, vars ...string) (
	map[string][]byte, *rpctypes.ResultABCIQuery, error,
) {
	var rqb []byte
	var err error
	if len(vars) > 0 {
		rqb, err = query.SysvarsRequest(vars).MarshalMsg(nil)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to marshal sysvar request")
		}
	}
	// perform the query
	res, err := node.ABCIQuery(query.SysvarsEndpoint, rqb)
	if err != nil {
		return nil, res, err
	}
	resp := make(query.SysvarsResponse)
	_, err = resp.UnmarshalMsg(res.Response.Value)
	return resp, res, err
}

// Sysvar gets a single system variable given its name and an example of its type
//
// The example is populated with the appropriate data
func Sysvar(node client.ABCIClient, name string, example msgp.Unmarshaler) error {
	svs, _, err := Sysvars(node, name)
	if err != nil {
		return err
	}
	svb, ok := svs[name]
	if !ok {
		return errors.New("node did not return sysvar " + name)
	}
	_, err = example.UnmarshalMsg(svb)
	return err
}

// SysvarHistory gets the value history of the given sysvar.
// Pass in 0,0 for the paging params to get the entire history.
func SysvarHistory(
	node client.ABCIClient,
	name string,
	after uint64,
	limit int,
) (*query.SysvarHistoryResponse, *rpctypes.ResultABCIQuery, error) {
	params := search.SysvarHistoryParams{
		Name:        name,
		AfterHeight: after,
		Limit:       limit,
	}

	paramsBuf, err := json.Marshal(params)
	if err != nil {
		return nil, nil, err
	}

	res, err := node.ABCIQuery(query.SysvarHistoryEndpoint, paramsBuf)
	if err != nil {
		return nil, res, err
	}

	khr := new(query.SysvarHistoryResponse)
	_, err = khr.UnmarshalMsg(res.Response.Value)
	if err != nil {
		return nil, res, err
	}

	return khr, res, err
}

// SysvarsAsJSON converts a msgp-encoded sysvar map into a json-encoded one
//
// Either jsvs or err will always be nil on return, but never both.
func SysvarsAsJSON(svs map[string][]byte) (jsvs map[string]interface{}, err error) {
	jsvs = make(map[string]interface{})
	for name, sv := range svs {
		var buf bytes.Buffer
		_, err = msgp.UnmarshalAsJSON(&buf, sv)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshaling "+name)
		}
		var val interface{}
		bbytes := buf.Bytes()
		if len(bbytes) == 0 {
			jsvs[name] = ""
			continue
		}
		err = json.Unmarshal(bbytes, &val)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("converting %s to json", name))
		}
		jsvs[name] = val
	}
	return
}
