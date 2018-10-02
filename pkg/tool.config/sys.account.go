package config

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

// SysAccount stores data for a system account
//
// These are accounts supporting transactions like RFE, NNR:
// the address is stored as a system variable, but cached locally.
type SysAccount struct {
	Address address.Address        `toml:"address"`
	Keys    []signature.PrivateKey `toml:"keys"`
}
