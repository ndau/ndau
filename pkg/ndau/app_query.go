package ndau

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
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

	ad, exists := state.GetAccount(address, app.blockTime)
	response.Log = fmt.Sprintf("acct exists: %t", exists)
	ad.UpdateSettlements(app.blockTime)
	adBytes, err := ad.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "serializing account data")
		return
	}

	response.Value = adBytes
}
