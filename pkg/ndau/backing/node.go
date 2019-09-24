package backing

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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
	Active                 bool
	DistributionScript     []byte
	TMAddress              string
	Key                    signature.PublicKey
	managedVars            map[string]struct{}
	managedVarRegistration math.Timestamp
}

// IsActiveNode is true when the provided address is an active node
func (s State) IsActiveNode(node address.Address) bool {
	n, ok := s.Nodes[node.String()]
	return ok && n.Active
}
