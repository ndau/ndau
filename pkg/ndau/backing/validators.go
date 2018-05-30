package backing

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	abci "github.com/tendermint/abci/types"

	util "github.com/oneiro-ndev/noms-util"
)

const validatorsKey = "validators"

// Validators are the map of public keys to powers in the validator set
type Validators nt.Map

// UpdateValidator updates the app's internal state with the given validator
func (state *State) UpdateValidator(db datas.Database, v abci.Validator) {
	validators := nt.Struct(*state).Get(validatorsKey).(nt.Map)
	pkBlob := util.Blob(db, v.GetPubKey())
	if v.Power == 0 {
		validators = validators.Edit().Remove(pkBlob).Map()
	} else {
		powerBlob := util.Int(v.Power).ToBlob(db)
		validators = validators.Edit().Set(pkBlob, powerBlob).Map()
	}
	*state = State(nt.Struct(*state).Set(validatorsKey, validators))
}
