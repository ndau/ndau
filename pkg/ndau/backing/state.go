package backing

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"

	meta "github.com/oneiro-ndev/metanode/pkg/meta.app"
)

const stateName = "State"

// State is a noms struct containing all Ndau chain state.
type State nt.Struct

// make sure State is a metaapp.State
var _ meta.State = (*State)(nil)

func newState(db datas.Database) nt.Struct {
	return nt.NewStruct(stateName, map[string]nt.Value{
		// Validators is a map of public key to power
		validatorsKey: nt.NewMap(db),
		// Accounts is a map of Address to AccountData
		accountsKey: nt.NewMap(db),
	})
}

// NewState initializes the Ndau chain's state
func NewState(db datas.Database) State {
	return State(newState(db))
}

// LoadState gets the application state from a dataset
//
// If it does not exist, a new state is automatically created
func LoadState(db datas.Database, ds datas.Dataset) (State, datas.Dataset, error) {
	var err error
	head, hasHead := ds.MaybeHeadValue()
	if !hasHead {
		head = newState(db)
		// commit the empty head so when we go to get things later, we don't
		// panic due to an empty dataset
		ds, err = db.CommitValue(ds, head)
		if err != nil {
			return State{}, ds, errors.Wrap(err, "LoadState failed to commit new head")
		}
	}
	nsS, isS := head.(nt.Struct)
	if !isS {
		return NewState(db), ds, errors.New("LoadState found non-struct as ds.HeadValue")
	}

	return State(nsS), ds, nil
}

// Load a state from a DB and DS, satisfying meta.State
func (state *State) Load(db datas.Database, ds datas.Dataset) (datas.Dataset, error) {
	var err error
	*state, ds, err = LoadState(db, ds)
	return ds, err
}

// Commit the current state and return an updated dataset
func (state *State) Commit(db datas.Database, ds datas.Dataset) (datas.Dataset, error) {
	return db.CommitValue(ds, nt.Struct(*state))
}
