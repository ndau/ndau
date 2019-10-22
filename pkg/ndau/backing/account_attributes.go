package backing

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

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

	progenitor := &addr
	acct, exists := s.Accounts[addr.String()]
	if exists && acct.Progenitor != nil {
		progenitor = acct.Progenitor
	}

	if attributes, ok := accountAttributes[progenitor.String()]; ok {
		if _, ok := attributes[attr]; ok {
			return true, nil
		}
	}

	return false, nil
}
