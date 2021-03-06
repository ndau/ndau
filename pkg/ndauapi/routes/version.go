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
	"regexp"

	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndau/pkg/tool"
)

// VersionResult is returned from the /version request; it retrieves
// information from the ndau chain
type VersionResult struct {
	NdauVersion string
	NdauSha     string
	Network     string
}

// HandleVersion is an http handler for version info
func HandleVersion(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version, _, err := tool.Version(cf.Node)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving version info", err, http.StatusInternalServerError))
			return
		}

		if version == "" {
			version = "unidentified-0-gunknown"
		}

		// The network name is stored in the tendermint chainid.
		// It's expected to be used to identify the network that the API is talking to,
		// as a way to differentiate the ndau mainnet from test networks.
		block, err := cf.Node.Block(nil)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving block info", err, http.StatusInternalServerError))
		}
		networkName := "unknown"
		if block != nil && block.Block != nil {
			networkName = block.Block.ChainID
		}

		result := VersionResult{
			NdauVersion: version,
			Network:     networkName,
		}

		// Our default version format is generated by "git describe --long --tags"
		// which looks like "v0.7.8-23-g7c8eac5", where the 3 parts separated by dashes
		// are the version tag, the number of commits since then, and "g" followed by the
		// current commit hash.
		// We don't care about the number of commits, and the g is not interesting, so
		// we pattern match the middle part.
		// If the format doesn't match that, we will just return the version string unmodified.
		p := regexp.MustCompile("-[0-9]+-g")
		spv := p.Split(version, -1)
		result.NdauVersion = spv[0]
		if len(spv) > 1 {
			result.NdauSha = spv[len(spv)-1]
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: http.StatusOK})
	}
}
