package config

import (
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// An Account contains the data necessary to interact with an account:
//
// ownership keys, transfer keys if assigned, an account nickname, and an address
type Account struct {
	Name             string          `toml:"name"`
	Address          address.Address `toml:"address"`
	Root             Keypair         `toml:"root"`
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
func (a *Account) TransferPrivateK(keys int) []signature.PrivateKey {
	return FilterK(a.TransferPrivate(), keys)
}

// FilterK filters a list of private keys by k, treating k as a bitset.
//
// Keys appear in the output list if their index in the input list corresponds
// with a 1 in the bit index of k.
func FilterK(keys []signature.PrivateKey, k int) []signature.PrivateKey {
	if k < 0 {
		return keys
	}

	out := make([]signature.PrivateKey, 0, len(keys))
	for k > 0 && len(keys) > 0 {
		pk := keys[0]
		keys = keys[1:]
		use := k&1 > 0
		k = k >> 1

		if use {
			out = append(out, pk)
		}
	}
	return out
}

func (a *Account) highestTransferPath() (uint64, uint64) {
	var high3, high4 uint64
	for _, tr := range a.Transfer {
		if tr.Path != nil {
			var field3, field4 uint64
			n, err := fmt.Sscanf(*tr.Path, AccountPathFormat, &field3, &field4)
			if err != nil {
				continue
			}
			if n != 2 {
				continue
			}
			if field3 > high3 || (field3 == high3 && field4 > high4) {
				high3 = field3
				high4 = field4
			}
		}
	}
	return high3, high4
}

func (a *Account) nextTransferPath() *string {
	if a.Ownership.Path == nil {
		return nil
	}
	high3, high4 := a.highestTransferPath()
	if (high3 == 0 || high3 == AccountListOffset) && (high4 == 0 || high4 == AccountStartNumber) {
		h := fmt.Sprintf(AccountPathFormat, TransferKeyOffset, AccountStartNumber)
		return &h
	}
	h := fmt.Sprintf(AccountPathFormat, high3, high4+1)
	return &h
}

// MakeTransferKey makes a transfer key for this account
//
// It does not actually add it to the keys list--that may be contraindicated
// by errors further on.
func (a *Account) MakeTransferKey(path *string) (newKeys *Keypair, err error) {
	newKeys = &Keypair{}
	if a.Ownership.Path == nil {
		// it's probably a non-hd key, so just proceed on that assumption
		newKeys.Public, newKeys.Private, err = signature.Generate(signature.Ed25519, nil)
		if err != nil {
			return nil, errors.Wrap(err, "generating new transfer key")
		}
	} else {
		// probably HD
		ekey, err := key.FromSignatureKey(&a.Ownership.Private)
		if err != nil {
			return nil, errors.Wrap(err, "extracting hd ownership key")
		}
		if path == nil || len(*path) == 0 {
			path = a.nextTransferPath()
		}
		if path == nil {
			return nil, errors.New("could not compute next transfer path")
		}

		newKeys.Path = path

		prive, err := ekey.DeriveFrom(*a.Root.Path, *path)
		if err != nil {
			return nil, errors.Wrap(err, "deriving child private key")
		}
		privs, err := prive.SPrivKey()
		if err != nil {
			return nil, errors.Wrap(err, "converting child private key to ndau fmt")
		}
		newKeys.Private = *privs

		pube, err := prive.Public()
		if err != nil {
			return nil, errors.Wrap(err, "converting child private key to public")
		}
		pubs, err := pube.SPubKey()
		if err != nil {
			return nil, errors.Wrap(err, "converting child public key to ndau fmt")
		}
		newKeys.Public = *pubs
	}

	return
}
