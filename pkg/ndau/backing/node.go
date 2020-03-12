package backing

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
)

//go:generate msgp -io=0

// Node keeps track of nodes in the validator and verifier sets
//
// Types here are noms-compatible for ease of marshalling and unmarshalling;
// though they're public for auto-marshalling purposes, they're not really
// meant for public access. Instead, the intent is that helper functions
// will manage all changes and handle type conversions.
//
//nomsify Node
type Node struct {
	Active                 bool                `json:"active"`
	DistributionScript     []byte              `json:"distribution_script"`
	TMAddress              string              `json:"tm_address"`
	Key                    signature.PublicKey `json:"public_key"`
	managedVars            map[string]struct{}
	managedVarRegistration math.Timestamp
}

// IsActiveNode is true when the provided address is an active node
func (s State) IsActiveNode(node address.Address) bool {
	n, ok := s.Nodes[node.String()]
	return ok && n.Active
}
