package generator

import (
	"errors"
	"fmt"
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

func parseType(t ast.Expr) (string, error) {
	switch defType := t.(type) {
	case *ast.Ident:
		return defType.Name, nil
	case *ast.SelectorExpr:
		xid, ok := defType.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf(
				"malformed selector: not ident @ [%d:%d]",
				defType.X.Pos(),
				defType.X.End(),
			)
		}
		return fmt.Sprintf("%s.%s", xid.Name, defType.Sel.Name), nil
	case *ast.ArrayType:
		r, err := parseType(defType.Elt)
		return "[]" + r, err
	default:
		return "", fmt.Errorf(
			"unknown type @ [%d:%d]",
			t.Pos(),
			t.End(),
		)
	}
}

// ParseField parses the given node of the AST as if it were a Field
//
// Because an AST field might define multiple logical fields,
// this may return multiple values
func ParseField(f *ast.Field) ([]Field, error) {
	fieldType, err := parseType(f.Type)
	if err != nil {
		return nil, err
	}

	out := make([]Field, 0, len(f.Names))
	for _, ident := range f.Names {
		field := NewField(ident.Name, fieldType)
		out = append(out, field)
	}
	return out, err
}

// ParseTransaction parses the given node of the AST as if it were a Transaction
func ParseTransaction(name string, node ast.Node) (out Transaction, err error) {
	s, ok := node.(*ast.StructType)
	if !ok {
		err = errors.New("node must be a struct definition")
		return
	}

	out.Name = name
	out.Comment = fmt.Sprintf("%s is a mobile compatible wrapper for a %s transaction", name, name)
	// TODO: docs
	// Parsing go docs is kind of terrible, and not critically important, so we
	// skip it for now. See: https://stackoverflow.com/a/19580785/504550

	var parsedFields []Field
	for _, field := range s.Fields.List {
		parsedFields, err = ParseField(field)
		if err != nil {
			return
		}

		out.Fields = append(out.Fields, parsedFields...)
	}

	return
}
