package backing

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// AccountHasAttribute returns whether the account with the given address has
// the given account attribute.
// Valid attributes can be found in system_vars/pkg/system_vars/account_attributes.go
func (s *State) AccountHasAttribute(addr address.Address, attr string) (bool, error) {
	aab, ok := s.Sysvars[sv.AccountAttributesName]
	if !ok {
		// if the sysvar is not set, the account doesn't have the attribute
		return false, nil
	}

	// unpack the struct
	accountAttributes := sv.AccountAttributes{}
	_, err := accountAttributes.UnmarshalMsg(aab)
	if err != nil {
		return false, errors.Wrap(err, "could not get AccountAttributes system variable")
	}

	acct, exists := s.Accounts[addr.String()]
	if !exists {
		return false, nil
	}

	var progenitor *address.Address
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