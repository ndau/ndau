package sdk

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Prevalidate prevalidates the provided transactable
func (c *Client) Prevalidate(tx metatx.Transactable) (fee math.Ndau, sib math.Ndau, err error) {
	result := new(routes.PrevalidateResult)
	err = c.post(tx, result, c.URL("tx/prevalidate/%s", metatx.NameOf(tx)))
	if err != nil {
		err = errors.Wrap(err, "prevalidating")
		return
	}
	fee = math.Ndau(result.FeeNapu)
	sib = math.Ndau(result.SibNapu)
	return
}

// Prevalidate prevalidates the provided transactable
func Prevalidate(node *Client, tx metatx.Transactable) (fee math.Ndau, sib math.Ndau, err error) {
	return node.Prevalidate(tx)
}

// Send broadcasts and commits a transaction
func (c *Client) Send(tx metatx.Transactable) (result *routes.SubmitResult, err error) {
	err = c.post(tx, result, c.URL("tx/submit/%s", metatx.NameOf(tx)))
	err = errors.Wrap(err, "submitting")
	return
}

// SendCommit broadcasts and commits a transaction
func SendCommit(node *Client, tx metatx.Transactable) (result *routes.SubmitResult, err error) {
	return node.Send(tx)
}
