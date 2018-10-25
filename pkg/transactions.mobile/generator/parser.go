package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// TransactionPath is the unexpanded path to the transaction definition file
const TransactionPath = "$GOPATH/src/github.com/oneiro-ndev/ndau/pkg/ndau/transactions.go"

// ParseTransactions parses the transaction definition file
func ParseTransactions() (*ast.File, error) {
	fp := os.ExpandEnv(TransactionPath)
	fset := token.NewFileSet()
	return parser.ParseFile(fset, fp, nil, parser.ParseComments)
}
