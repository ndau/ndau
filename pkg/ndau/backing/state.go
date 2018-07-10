package backing

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	meta "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

const accountKey = "accounts"
const delegateKey = "delegates"
const bpcSequenceKey = "bpcSequences"

// State is primarily a set of accounts
type State struct {
	Accounts map[string]AccountData
	// Delegates is a map of strings to a set of strings
	// All strings are addresses
	// the keys are the delegated nodes
	// the values are the addresses of the accounts which those nodes must
	// compute
	Delegates map[string]map[string]struct{}

	// The BPC is likely to have several transactions, for which we'll want
	// independent sequence numbers, as they'll be delegated to different
	// individuals and groups. We keep them here.
	BPCSequences struct {
		ReleaseFromEndowment util.Int
	}
}

// make sure State is a metaapp.State
var _ meta.State = (*State)(nil)

// Init satisfies meta.State
func (s *State) Init(nt.ValueReadWriter) {
	s.Accounts = make(map[string]AccountData)
	s.Delegates = make(map[string]map[string]struct{})
}

// MarshalNoms satisfies noms' Marshaler interface
func (s State) MarshalNoms(vrw nt.ValueReadWriter) (nt.Value, error) {
	ns := nt.NewStruct("state", nt.StructData{
		accountKey:  nt.NewMap(vrw),
		delegateKey: nt.NewMap(vrw),
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

	// marshal bpc sequences
	bpcSeq, err := marshal.Marshal(vrw, s.BPCSequences)
	if err != nil {
		return nil, errors.Wrap(err, "Marshaling BPC sequences")
	}
	ns = ns.Set(bpcSequenceKey, bpcSeq)

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

	// unmarshal sequences
	err = marshal.Unmarshal(st.Get(bpcSequenceKey), s.BPCSequences)
	err = errors.Wrap(err, "Unmarshaling BPC sequences")
	return err
}
