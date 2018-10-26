package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// ParseTransactions parses the transaction definition file
func ParseTransactions() (*ast.File, error) {
	fp := os.ExpandEnv(TransactionPath)
	fset := token.NewFileSet()
	return parser.ParseFile(fset, fp, nil, parser.ParseComments)
}
