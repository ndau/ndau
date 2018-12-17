package txgen

import (
	"errors"
	"go/ast"
)

// TxIDs is the name of the transaction IDs map
const TxIDs = "TxIDs"

// GetTxNames gets the names of the transactions from the TxIDs map
func GetTxNames(n ast.Node) ([]string, error) {
	c, ok := n.(*ast.CompositeLit)
	if !ok {
		return nil, errors.New("n must be a composite literal")
	}
	if _, ok := c.Type.(*ast.MapType); !ok {
		return nil, errors.New("n must be a map")
	}

	out := make([]string, 0, len(c.Elts))

	for _, elt := range c.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			return nil, errors.New("elt must be a key-value expr")
		}

		ref, ok := kv.Value.(*ast.UnaryExpr)
		if !ok {
			return nil, errors.New("elt value must be a reference")
		}

		lit, ok := ref.X.(*ast.CompositeLit)
		if !ok {
			return nil, errors.New("elt value must be a literal")
		}

		ident, ok := lit.Type.(*ast.Ident)
		if !ok {
			return nil, errors.New("elt value must have an ident type")
		}

		out = append(out, ident.Name)
	}

	return out, nil
}
