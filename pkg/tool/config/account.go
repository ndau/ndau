package config

import (
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

// An Account contains the data necessary to interact with an account:
//
// ownership keys, transfer keys if assigned, an account nickname, and an address
type Account struct {
	Name      string
	Address   address.Address
	Ownership Keypair
	Transfer  *Keypair
}

func (acct Account) toToml() (tomlAccount, error) {
	ownership, err := acct.Ownership.toToml()
	if err != nil {
		return tomlAccount{}, errors.Wrap(err, "ownership")
	}
	var transfer *tomlKeypair
	if acct.Transfer != nil {
		tr, err := acct.Transfer.toToml()
		if err != nil {
			return tomlAccount{}, errors.Wrap(err, "transfer")
		}
		transfer = &tr
	}
	return tomlAccount{
		Name:      acct.Name,
		Address:   acct.Address.String(),
		Ownership: ownership,
		Transfer:  transfer,
	}, nil
}

// String satisfies io.Stringer
func (acct Account) String() string {
	var id string
	if acct.Name != "" {
		id = acct.Name
	} else {
		id = acct.Address.String()
	}
	return fmt.Sprintf("%s: owner %s transfer %s", id, acct.Ownership, acct.Transfer)
}

// tomlAccount is an account being prepared for toml marshaling
type tomlAccount struct {
	Name      string       `toml:"name"`
	Address   string       `toml:"address"`
	Ownership tomlKeypair  `toml:"ownership"`
	Transfer  *tomlKeypair `toml:"transfer"`
}

func (ta tomlAccount) toAccount() (Account, error) {
	ownership, err := ta.Ownership.toKeypair()
	if err != nil {
		return Account{}, errors.Wrap(err, "ownership")
	}

	var transfer *Keypair
	if ta.Transfer != nil {
		tr, err := ta.Transfer.toKeypair()
		if err != nil {
			return Account{}, errors.Wrap(err, "transfer")
		}
		transfer = &tr
	}

	addr, err := address.Validate(ta.Address)
	if err != nil {
		return Account{}, errors.Wrap(err, "address")
	}

	return Account{
		Name:      ta.Name,
		Address:   addr,
		Ownership: ownership,
		Transfer:  transfer,
	}, nil
}
