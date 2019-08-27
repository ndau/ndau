package sdk

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

// GetCurrencySeats gets a list of ndau currency seats
func (c *Client) GetCurrencySeats() (seats []address.Address, err error) {
	err = c.get(seats, "account/currencyseats")
	err = errors.Wrap(err, "getting currency seats from API")
	return
}

// GetCurrencySeats gets a list of ndau currency seats
func GetCurrencySeats(node *Client) ([]address.Address, error) {
	return node.GetCurrencySeats()
}

// GetDelegates gets the set of nodes with delegates, and the list of accounts delegated to each
func (c *Client) GetDelegates() (delegates map[address.Address][]address.Address, err error) {
	err = c.get(delegates, "state/delegates")
	err = errors.Wrap(err, "getting delegates from API")
	return
}

// GetDelegates gets the set of nodes with delegates, and the list of accounts delegated to each
func GetDelegates(node *Client) (map[address.Address][]address.Address, error) {
	return node.GetDelegates()
}
