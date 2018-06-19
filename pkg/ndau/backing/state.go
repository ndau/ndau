package backing

import (
	"errors"

	nt "github.com/attic-labs/noms/go/types"
	meta "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
)

// State is primarily a set of accounts
type State struct {
	Accounts map[string]AccountData
}

// make sure State is a metaapp.State
var _ meta.State = (*State)(nil)

// Init satisfies meta.State
func (s *State) Init(nt.ValueReadWriter) {
	s.Accounts = make(map[string]AccountData)
}

// MarshalNoms satisfies noms' Marshaler interface
func (s State) MarshalNoms(vrw nt.ValueReadWriter) (nt.Value, error) {
	nm := nt.NewMap(vrw)
	editor := nm.Edit()
	for k, v := range s.Accounts {
		vval, err := v.MarshalNoms(vrw)
		if err != nil {
			return nm, err
		}
		editor.Set(nt.String(k), vval)
	}
	return editor.Map(), nil
}

// UnmarshalNoms satisfies noms' Unmarshaler interface
func (s *State) UnmarshalNoms(v nt.Value) (err error) {
	m, isMap := v.(nt.Map)
	if !isMap {
		return errors.New("v is not an nt.Map")
	}
	nm := make(map[string]AccountData)

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
					nm[string(k)] = ad
				}
			}
		}
	})
	if err == nil {
		s.Accounts = nm
	}
	return err
}
