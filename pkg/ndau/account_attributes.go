package ndau

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// Return whether the account with the given address has the given account attribute.
// Valid attributes can be found in system_vars/pkg/system_vars/account_attributes.go
func (app *App) accountHasAttribute(addr address.Address, attr string) (bool, error) {
	accountAttributes := sv.AccountAttributes{}
	err := app.System(sv.AccountAttributesName, &accountAttributes)
	if err != nil {
		return false, err
	}

	var progenitor *address.Address
	acct, _ := app.getAccount(addr)
	if acct.Progenitor == nil {
		progenitor = &addr
	} else {
		progenitor = acct.Progenitor
	}

	if attributes, ok := accountAttributes[progenitor.String()]; ok {
		if _, ok := attributes[attr]; ok {
			return true, nil
		}
	}

	return false, nil
}
