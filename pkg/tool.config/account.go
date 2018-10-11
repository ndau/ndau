package config

import (
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// An Account contains the data necessary to interact with an account:
//
// ownership keys, transfer keys if assigned, an account nickname, and an address
type Account struct {
	Name             string          `toml:"name"`
	Address          address.Address `toml:"address"`
	Ownership        Keypair         `toml:"ownership"`
	Transfer         []Keypair       `toml:"transfer"`
	ValidationScript chaincode       `toml:"validation_script"`
}

func (a *Account) String() string {
	return fmt.Sprintf(
		"%s: %s...%s (%d tr keys)",
		a.Name,
		a.Address.String()[:8],
		a.Address.String()[len(a.Address.String())-5:],
		len(a.Transfer),
	)
}

// TransferPublic constructs a list of all private transfer keys
func (a *Account) TransferPublic() []signature.PublicKey {
	pks := make([]signature.PublicKey, 0, len(a.Transfer))
	for _, kp := range a.Transfer {
		pks = append(pks, kp.Public)
	}
	return pks
}

// TransferPrivate constructs a list of all private transfer keys
func (a *Account) TransferPrivate() []signature.PrivateKey {
	pks := make([]signature.PrivateKey, 0, len(a.Transfer))
	for _, kp := range a.Transfer {
		pks = append(pks, kp.Private)
	}
	return pks
}

// TransferPrivateK constructs a list of all private transfer keys which have
// their bits set, treating `keys` as a bitset with the lowest bit corresponding
// to the 0 index of the list of private keys.
func (a *Account) TransferPrivateK(keys *int) []signature.PrivateKey {
	return FilterK(a.TransferPrivate(), keys)
}

// FilterK filters a list of private keys by k, treating k as a bitset.
//
// Keys appear in the output list if their index in the input list corresponds
// with a 1 in the bit index of k.
func FilterK(keys []signature.PrivateKey, k *int) []signature.PrivateKey {
	if k == nil || *k <= 0 {
		return keys
	}

	out := make([]signature.PrivateKey, 0, len(keys))
	for *k > 0 && len(keys) > 0 {
		pk := keys[0]
		keys = keys[1:]
		use := *k&1 > 0
		*k = *k >> 1

		if use {
			out = append(out, pk)
		}
	}
	return out
}
