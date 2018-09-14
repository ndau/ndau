package backing

import (
	nt "github.com/attic-labs/noms/go/types"
	meta "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

const accountKey = "accounts"
const delegateKey = "delegates"
const nodeKey = "nodes"
const lnrnKey = "lnrn" // last node reward nomination
const pnrKey = "pnr"   // pending node reward
const unrKey = "unr"   // uncredited node reward

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
	// The last node reward nomination is necessary in state because this
	// governs the validity of upcoming node reward nominations; there's
	// a minimum interval between them.
	LastNodeRewardNomination math.Timestamp
	// Node rewards are a bit complex. They're accumulated with every
	// CreditEAI transaction into the PendingNodeReward variable. On
	// NominateNodeReward, the balance in PendingNodeReward is moved into
	// UnclaimedNodeReward, because there may be further CreditEAI transactions,
	// which have to be stored up for the subsequent node.
	// ClaimNodeReward transactions actually claim the unclaimed node reward;
	// otherwise, it's overwritten at the next Nominate tx.
	PendingNodeReward   math.Ndau
	UnclaimedNodeReward math.Ndau
	// Of course, we have to keep track of which node has acutally won
	NodeRewardWinner address.Address
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
		lnrnKey:     util.Int(0).ToBlob(vrw),
		pnrKey:      util.Int(0).ToBlob(vrw),
		unrKey:      util.Int(0).ToBlob(vrw),
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
	nm, err := MarshalNodesNoms(vrw, s.Nodes)
	if err != nil {
		return ns, err
	}
	ns = ns.Set(nodeKey, nm)
	// marshal last node reward nomination
	ns = ns.Set(
		lnrnKey,
		util.Int(s.LastNodeRewardNomination).ToBlob(vrw),
	)
	// marshal pending node reward
	ns = ns.Set(
		pnrKey,
		util.Int(s.PendingNodeReward).ToBlob(vrw),
	)
	// marshal unclaimed node reward
	ns = ns.Set(
		unrKey,
		util.Int(s.UnclaimedNodeReward).ToBlob(vrw),
	)

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
	s.Nodes, err = UnmarshalNodesNoms(st.Get(nodeKey))
	if err != nil {
		return errors.Wrap(err, "unmarshalling nodes")
	}

	// unmarshal last node reward nomination
	lnrnI, err := util.IntFromBlob(st.Get(lnrnKey).(nt.Blob))
	if err != nil {
		return errors.Wrap(err, "unmarshalling last node reward nomination")
	}
	s.LastNodeRewardNomination = math.Timestamp(lnrnI)
	// unmarshal pending node reward
	pnrI, err := util.IntFromBlob(st.Get(pnrKey).(nt.Blob))
	if err != nil {
		return errors.Wrap(err, "unmarshalling pending node reward")
	}
	s.PendingNodeReward = math.Ndau(pnrI)
	// unmarshal unclaimed node reward
	unrI, err := util.IntFromBlob(st.Get(unrKey).(nt.Blob))
	if err != nil {
		return errors.Wrap(err, "unmarshalling unclaimed node reward")
	}
	s.UnclaimedNodeReward = math.Ndau(unrI)

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
