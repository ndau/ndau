package address

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	util "github.com/oneiro-ndev/noms-util"
)

// Address is the serialized representation of an address.Data struct
type Address []byte

// static assertions that Address satisfies Marshaler and Unmarshaler
var _ marshal.Marshaler = (*Address)(nil)
var _ marshal.Unmarshaler = (*Address)(nil)

// MarshalNoms satisfies marshal.Marshaler
func (a Address) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return util.Blob(vrw, a), nil
}

// UnmarshalNoms satisfies marshal.Unmarshaler
func (a *Address) UnmarshalNoms(v nt.Value) (err error) {
	*a, err = util.Unblob(v.(nt.Blob))
	return
}
