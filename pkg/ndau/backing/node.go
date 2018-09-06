package backing

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	util "github.com/oneiro-ndev/noms-util"
)

// Node keeps track of nodes in the validator and verifier sets
//
// Types here are noms-compatible for ease of marshalling and unmarshalling;
// though they're public for auto-marshalling purposes, they're not really
// meant for public access. Instead, the intent is that helper functions
// will manage all changes and handle type conversions.
type Node struct {
	TotalStake util.Int
	Costakers  map[string]util.Int
}

// NewNode creates a new node from its self-stake
func NewNode(addr address.Address, stake math.Ndau) Node {
	return Node{
		TotalStake: util.Int(stake),
		Costakers: map[string]util.Int{
			addr.String(): util.Int(stake),
		},
	}
}

// Costake adds a costaker to a node
func (n *Node) Costake(addr address.Address, stake math.Ndau) {
	addrS := addr.String()
	// if this is already a costaker, nop
	_, isCostaker := n.Costakers[addrS]
	if isCostaker {
		return
	}

	n.TotalStake += util.Int(stake)
	n.Costakers[addrS] = util.Int(stake)
}

// Unstake removes a costaker from a node
func (n *Node) Unstake(addr address.Address) {
	addrS := addr.String()
	// if this wasn't already a costaker, staked == 0, so just continue
	staked := n.Costakers[addrS]
	delete(n.Costakers, addrS)
	n.TotalStake -= staked
}
