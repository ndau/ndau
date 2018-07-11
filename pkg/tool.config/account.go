package config

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

// An Account contains the data necessary to interact with an account:
//
// ownership keys, transfer keys if assigned, an account nickname, and an address
type Account struct {
	Name      string          `toml:"name"`
	Address   address.Address `toml:"address"`
	Ownership Keypair         `toml:"ownership"`
	Transfer  *Keypair        `toml:"transfer"`
}
