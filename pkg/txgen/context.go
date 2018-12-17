package txgen

import (
	"fmt"

	"github.com/oneiro-ndev/ndau/pkg/txmobile/generator"
	"github.com/pkg/errors"
)

// Context defines all the info the template needs
type Context struct {
	Transactions []Transaction
}

// MakeContext makes a context object from ndau transactions
func MakeContext() (*Context, error) {
	ast, err := ParseTransactions()
	if err != nil {
		return nil, errors.Wrap(err, "parsing transactions")
	}

	txIDs := FindDefinition(ast, generator.TxIDs)
	if txIDs == nil {
		return nil, errors.New("TxIDs not found")
	}

	txNames, err := generator.GetTxNames(txIDs.Definition)
	if err != nil {
		return nil, errors.Wrap(err, "getting tx names")
	}

	transactions := make([]Transaction, 0, len(txNames))

	for _, n := range txNames {
		def := generator.FindDefinition(ast, n)
		if def == nil {
			return nil, fmt.Errorf("tx %s not found", n)
		}

		transaction, err := ParseTransaction(n, def.Definition)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parsing %s tx", n))
		}

		transactions = append(transactions, transaction)
	}

	return &Context{
		Transactions: transactions,
	}, nil
}
