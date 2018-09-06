package backing

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	meta "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

const accountKey = "accounts"
const delegateKey = "delegates"
const nodeKey = "nodes"

// State is primarily a set of accounts
type State struct {
	Accounts map[string]AccountData
	// Delegates is a map of strings to a set of strings
	// All strings are addresses
	// the keys are the delegated nodes
	// the values are the addresses of the accounts which those nodes must
	// compute
	Delegates map[string]map[string]struct{}
	// Nodes keeps track of the validator and verifier node stakes.
	// The key is the node address. The value is a Node struct.
	Nodes map[string]Node
}

// make sure State is a metaapp.State
var _ meta.State = (*State)(nil)

// Init satisfies meta.State
func (s *State) Init(nt.ValueReadWriter) {
	s.Accounts = make(map[string]AccountData)
	s.Delegates = make(map[string]map[string]struct{})
	s.Nodes = make(map[string]Node)
}

// MarshalNoms satisfies noms' Marshaler interface
func (s State) MarshalNoms(vrw nt.ValueReadWriter) (nt.Value, error) {
	ns := nt.NewStruct("state", nt.StructData{
		accountKey:  nt.NewMap(vrw),
		delegateKey: nt.NewMap(vrw),
		nodeKey:     nt.NewMap(vrw),
	})

	// marshal accounts
	editor := ns.Get(accountKey).(nt.Map).Edit()
	for k, v := range s.Accounts {
		vval, err := v.MarshalNoms(vrw)
		if err != nil {
			return ns, err
		}
		editor.Set(nt.String(k), vval)
	}
	ns = ns.Set(accountKey, editor.Map())
	// marshal delegates
	editor = ns.Get(delegateKey).(nt.Map).Edit()
	for delegateNode, delegateAddresses := range s.Delegates {
		daSet := nt.NewSet(vrw)
		setEditor := daSet.Edit()
		for delegateAddress := range delegateAddresses {
			setEditor.Insert(nt.String(delegateAddress))
		}
		editor.Set(nt.String(delegateNode), setEditor.Set())
	}
	ns = ns.Set(delegateKey, editor.Map())
	// marshal nodes
	nm, err := marshal.Marshal(vrw, s.Nodes)
	if err != nil {
		return ns, err
	}
	ns = ns.Set(nodeKey, nm)

	return ns, nil
}

// UnmarshalNoms satisfies noms' Unmarshaler interface
func (s *State) UnmarshalNoms(v nt.Value) (err error) {
	st, isStruct := v.(nt.Struct)
	if !isStruct {
		return errors.New("v is not an nt.Struct")
	}

	// unmarshal accounts
	mV := st.Get(accountKey)
	m, isMap := mV.(nt.Map)
	if !isMap {
		return errors.New("account data not a nt.Map")
	}
	s.Accounts = make(map[string]AccountData)
	m.IterAll(func(key, value nt.Value) {
		if err == nil {
			ad := AccountData{}
			err = ad.UnmarshalNoms(value)
			if err == nil {
				k, keyIsString := key.(nt.String)
				if !keyIsString {
					err = errors.New("non-nt.String key")
				}
				if err == nil {
					s.Accounts[string(k)] = ad
				}
			}
		}
	})
	if err != nil {
		return err
	}

	// unmarshal delegates
	mV = st.Get(delegateKey)
	m, isMap = mV.(nt.Map)
	if !isMap {
		return errors.New("delegates not a nt.Map")
	}
	s.Delegates = make(map[string]map[string]struct{})
	m.IterAll(func(key, value nt.Value) {
		if err == nil {
			inner := make(map[string]struct{})
			ks, isStr := key.(nt.String)
			if !isStr {
				err = errors.New("delegates key not nt.String")
			}
			if err == nil {
				innerSet, isSet := value.(nt.Set)
				if !isSet {
					err = errors.New("delegates value is not nt.Set")
				}
				if err == nil {
					innerSet.IterAll(func(innerVal nt.Value) {
						if err == nil {
							setStr, isStr := innerVal.(nt.String)
							if !isStr {
								err = errors.New("delegates inner value is not nt.String")
							}
							if err == nil {
								inner[string(setStr)] = struct{}{}
							}
						}
					})
					s.Delegates[string(ks)] = inner
				}
			}
		}
	})

	// unmarshal nodes
	nV := st.Get(nodeKey)
	s.Nodes = make(map[string]Node)
	err = marshal.Unmarshal(nV, &s.Nodes)
	if err != nil {
		return errors.Wrap(err, "unmarshalling nodes")
	}

	return err
}

// GetAccount returns the account at the requested address.
//
// If the account does not already exist, a fresh one is created.
//
// This function is necessary because account zero values are not valid:
// the `Last*Update` fields must be initialized with the current block time.
//
// The boolean return value is true when the account previously existed;
// false when it is new.
func (s *State) GetAccount(address address.Address, blockTime math.Timestamp) (AccountData, bool) {
	data, hasAccount := s.Accounts[address.String()]
	if !hasAccount {
		data.LastEAIUpdate = blockTime
		data.LastWAAUpdate = blockTime
	}
	return data, hasAccount
}

// GetValidAccount returns a valid account at the requested address
//
// If the account does not already exist, a fresh one is created
//
// This is a sugar function to simplify some common validation requirements.
func (s *State) GetValidAccount(address address.Address, blockTime math.Timestamp, sequence uint64, signableBytes []byte, signatures []signature.Signature) (AccountData, bool, *bitset256.Bitset256, error) {
	ad, exists := s.GetAccount(address, blockTime)
	if sequence <= ad.Sequence {
		return ad, exists, nil, errors.New("Sequence too low")
	}
	validates, sigset := ad.ValidateSignatures(signableBytes, signatures)
	if !validates {
		return ad, exists, sigset, errors.New("Invalid signature(s)")
	}
	return ad, exists, sigset, nil
}

// Stake updates the state to handle staking an account to another
func (s *State) Stake(targetA, nodeA address.Address) error {
	nodeS := nodeA.String()
	node, isNode := s.Nodes[nodeS]
	// logically, the operation I want in this if is nxor, but go doesn't
	// define that for booleans, because reasons
	if (targetA == nodeA) == isNode {
		if isNode {
			return errors.New("cannot re-self-stake")
		}
		return errors.New("node is not already a node; can't stake to it")
	}

	target := s.Accounts[targetA.String()]
	if isNode {
		// targetA != nodeA
		node.Costake(targetA, target.Balance)
	} else {
		// targetA == nodeA
		node = NewNode(targetA, target.Balance)
	}

	s.Nodes[nodeS] = node
	return nil
}

// GetCostakers returns the list of costakers associated with a node
func (s *State) GetCostakers(nodeA address.Address) []AccountData {
	node, isNode := s.Nodes[nodeA.String()]
	if !isNode {
		return nil
	}

	out := make([]AccountData, 0, len(node.Costakers))
	for costaker := range node.Costakers {
		ad, hasAccount := s.Accounts[costaker]
		if hasAccount {
			out = append(out, ad)
		}
	}
	return out
}
