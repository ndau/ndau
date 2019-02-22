package backing

import (
	"bytes"
	"fmt"

	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

// Node keeps track of nodes in the validator and verifier sets
//
// Types here are noms-compatible for ease of marshalling and unmarshalling;
// though they're public for auto-marshalling purposes, they're not really
// meant for public access. Instead, the intent is that helper functions
// will manage all changes and handle type conversions.
type Node struct {
	Active             bool
	DistributionScript []byte
	RPCAddress         string
	TotalStake         math.Ndau
	Costakers          map[string]math.Ndau
}

// NewNode creates a new node from its self-stake
func NewNode(addr address.Address, stake math.Ndau) Node {
	return Node{
		TotalStake: stake,
		Costakers: map[string]math.Ndau{
			addr.String(): stake,
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

	n.TotalStake += stake
	n.Costakers[addrS] = stake
}

// Unstake removes a costaker from a node
func (n *Node) Unstake(addr address.Address) {
	addrS := addr.String()
	// if this wasn't already a costaker, staked == 0, so just continue
	staked := n.Costakers[addrS]
	delete(n.Costakers, addrS)
	n.TotalStake -= staked
}

var _ marshal.Marshaler = (*Node)(nil)
var _ marshal.Unmarshaler = (*Node)(nil)

// MarshalNoms implements marshal.Marshaler
func (n *Node) MarshalNoms(vrw nt.ValueReadWriter) (nt.Value, error) {
	// bytes.NewBuffer takes ownership of the passed slice, and may overwrite
	// the original data. We can't have that: we need to copy the data into
	// the buffer. To that end, we explicitly read the data into a new, empty buffer.
	dsBuffer := new(bytes.Buffer)
	dsBuffer.Write(n.DistributionScript)
	dsBlob := nt.NewBlob(vrw, dsBuffer)

	cNMapE := nt.NewMap(vrw).Edit()
	for costaker, stake := range n.Costakers {
		stakeV, err := util.Int(stake).MarshalNoms(vrw)
		if err != nil {
			return nil, fmt.Errorf("encoding stake (%d) to noms", stake)
		}
		cNMapE.Set(nt.String(costaker), stakeV)
	}

	return nt.NewStruct("node", nt.StructData{
		"active":             nt.Bool(n.Active),
		"distributionScript": dsBlob,
		"rpcAddress":         nt.String(n.RPCAddress),
		"totalStake":         util.Int(n.TotalStake).NomsValue(),
		"costakers":          cNMapE.Map(),
	}), nil
}

// UnmarshalNoms implements marshal.Unmarshaler
func (n *Node) UnmarshalNoms(v nt.Value) error {
	s, isS := v.(nt.Struct)
	if !isS {
		return errors.New("value is not nt.Struct")
	}
	n.Active = bool(s.Get("active").(nt.Bool))

	dsBuffer := new(bytes.Buffer)
	s.Get("distributionScript").(nt.Blob).Copy(dsBuffer)
	n.DistributionScript = dsBuffer.Bytes()

	n.RPCAddress = string(s.Get("rpcAddress").(nt.String))

	stakeI, err := util.IntFrom(s.Get("totalStake"))
	if err != nil {
		return errors.Wrap(err, "totalStake")
	}
	n.TotalStake = math.Ndau(stakeI)

	n.Costakers = make(map[string]math.Ndau)
	s.Get("costakers").(nt.Map).Iter(func(kV, vV nt.Value) (stop bool) {
		vI, err := util.IntFrom(vV)
		if err != nil {
			stop = true
			err = errors.Wrap(err, "parsing costaker stake from blob")
			return
		}
		n.Costakers[string(kV.(nt.String))] = math.Ndau(vI)
		return
	})
	return err
}

// MarshalNodesNoms marshals a map of nodes into a noms map
func MarshalNodesNoms(vrw nt.ValueReadWriter, in map[string]Node) (nt.Map, error) {
	outE := nt.NewMap(vrw).Edit()
	for addr, node := range in {
		nodeV, err := node.MarshalNoms(vrw)
		if err != nil {
			return outE.Map(), err
		}
		outE.Set(nt.String(addr), nodeV)
	}
	return outE.Map(), nil
}

// UnmarshalNodesNoms unmarshals a noms value into a map of nodes
func UnmarshalNodesNoms(in nt.Value) (map[string]Node, error) {
	out := make(map[string]Node)
	var iterErr error
	in.(nt.Map).Iter(func(kV, vV nt.Value) (stop bool) {
		k := string(kV.(nt.String))
		v := Node{}
		iterErr = v.UnmarshalNoms(vV)
		if iterErr != nil {
			return true
		}
		out[k] = v
		return
	})
	return out, iterErr
}
