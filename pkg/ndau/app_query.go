package ndau

import (
	abci "github.com/tendermint/tendermint/abci/types"

	meta "github.com/oneiro-ndev/metanode/pkg/meta.app"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
)

// AccountEndpoint is the endpoint at which Account queries live
const AccountEndpoint = "/account"

func init() {
	meta.RegisterQueryHandler(AccountEndpoint, accountQuery)
}

func accountQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	address, err := address.Validate(string(request.GetData()))
	if err != nil {
		app.QueryError(err, response, "deserializing address")
		return
	}

	state := app.GetState().(*backing.State)

	ad, exists := state.Accounts[address.String()]
	if exists {
		response.Log = "exists"
	} else {
		response.Log = "does not exist"
	}
	ad.UpdateEscrow(app.blockTime)
	adBytes, err := ad.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "serializing account data")
		return
	}

	response.Value = adBytes
}
