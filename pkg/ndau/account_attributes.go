package ndau

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Return whether the account with the given address has the given account attribute.
// Valid attributes can be found in system_vars/pkg/system_vars/account_attributes.go
func (app *App) accountHasAttribute(addr address.Address, attr string) (bool, error) {
	accountAttributes := sv.AccountAttributes{}
	exists, err := app.SystemOptional(sv.AccountAttributesName, &accountAttributes)
	if err != nil {
		if exists {
			// Some critical error occurred fetching the system variable.
			return false, errors.Wrap(err, "Could not fetch AccountAttributes system variable")
		}
		// The system variable doesn't exist, so no accounts have the given attribute.
		return false, nil
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
