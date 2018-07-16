package config

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

// RFEAuth contains the data necessary to issue a RFE transaction
//
// Address is the address of the account whose transfer key is in the
// list of authorized public RFE keys.
//
// Key is the private key associated with the transfer key of that address.
type RFEAuth struct {
	Address address.Address      `toml:"address"`
	Key     signature.PrivateKey `toml:"key"`
}
