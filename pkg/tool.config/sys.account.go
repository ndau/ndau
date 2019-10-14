package config

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// SysAccount stores data for a system account
//
// These are accounts supporting transactions like RFE, NNR:
// the address is stored as a system variable, but cached locally.
type SysAccount struct {
	Address address.Address        `toml:"address"`
	Keys    []signature.PrivateKey `toml:"keys"`
}

// SysAccountFromAssc creates a SysAccount from the associated data
func SysAccountFromAssc(assc generator.Associated, acct sv.SysAcct) (sa *SysAccount, err error) {
	addrI, addrOk := assc[acct.Address]
	privkeyI, keyOk := assc[acct.Validation.Private]
	if !(addrOk && keyOk) {
		err = fmt.Errorf(
			"assc: %s set: %t; %s set: %t",
			acct.Address, keyOk,
			acct.Validation.Private, keyOk,
		)
		return
	}

	addrS, addrOk := addrI.(string)
	if !addrOk {
		err = fmt.Errorf("assc: value of %s was not a string", acct.Address)
		return
	}
	privkeyS, keyOk := privkeyI.(string)
	if !keyOk {
		err = fmt.Errorf("assc: value of %s was not a string", acct.Validation.Private)
		return
	}

	var addr address.Address
	var privkey signature.PrivateKey

	addr, err = address.Validate(addrS)
	if err != nil {
		err = errors.Wrap(err, "validating address for "+acct.Name)
		return
	}

	err = privkey.UnmarshalText([]byte(privkeyS))
	if err != nil {
		err = errors.Wrap(err, "unmarshalling private validation key of "+acct.Name)
		return
	}

	sa = &SysAccount{
		Address: addr,
		Keys:    []signature.PrivateKey{privkey},
	}
	return
}
