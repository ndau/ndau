package backing

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"

	"github.com/oneiro-ndev/ndaunode/pkg/node.address"
)

const accountsKey = "accounts"

func (state *State) accounts() nt.Map {
	return nt.Struct(*state).Get(accountsKey).(nt.Map)
}

// GetAccount returns the AccountData struct for a given address
//
// If no account exists for that address, a default is created
func (state *State) GetAccount(db datas.Database, address address.Address) (ad AccountData, err error) {
	nomsAddress, err := address.MarshalNoms(db)
	if err != nil {
		err = errors.Wrap(err, "GetAccount failed to marshal address")
		return
	}

	nomsAd, hasAddress := state.accounts().MaybeGet(nomsAddress)
	if !hasAddress {
		return
	}

	err = ad.UnmarshalNoms(nomsAd)
	return
}

// UpdateAccount updates the app's account for the given address
func (state *State) UpdateAccount(db datas.Database, address address.Address, ad AccountData) error {
	nomsAddress, err := address.MarshalNoms(db)
	if err != nil {
		return errors.Wrap(err, "UpdateAccoutn failed to marshal address")
	}
	nomsAd, err := ad.MarshalNoms(db)
	if err != nil {
		return errors.Wrap(err, "UpdateAccount failed to marshal account data")
	}

	accounts := state.accounts().Edit().Set(nomsAddress, nomsAd).Map()
	*state = State(nt.Struct(*state).Set(accountsKey, accounts))
	return nil
}
