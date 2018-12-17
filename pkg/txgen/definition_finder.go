package txgen

import (
	"errors"
	"fmt"
	"go/ast"
	"io"

	"golang.org/x/tools/go/ast/astutil"
)

// DefinitionFinder finds definitions of variables by name within a node
type DefinitionFinder struct {
	Name           string
	Root           ast.Node
	DefinitionType string
	Definition     ast.Expr
}

// Visit implements ast.Visitor
func (t *DefinitionFinder) Visit(n ast.Node) ast.Visitor {
	switch node := n.(type) {
	case *ast.AssignStmt:
		for idx, name := range node.Lhs {
			if ident, ok := name.(*ast.Ident); ok {
				if ident.Name == t.Name {
					t.DefinitionType = fmt.Sprintf("%T", node)
					t.Definition = node.Rhs[idx]
					return nil
				}
			}
		}
	case *ast.GenDecl:
		for _, spec := range node.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				for vidx, name := range value.Names {
					if name.Name == t.Name {
						t.DefinitionType = fmt.Sprintf("%T", node)
						t.Definition = value.Values[vidx]
						return nil
					}
				}
			}
		}
	case *ast.TypeSpec:
		if node.Name.Name == t.Name {
			t.DefinitionType = fmt.Sprintf("%T", node)
			t.Definition = node.Type
			return nil
		}

	}
	return t
}

func (t *DefinitionFinder) Write(w io.Writer) error {
	if t.Definition == nil {
		return errors.New("definition not found")
	}
	f, ok := t.Root.(*ast.File)
	if !ok {
		return errors.New("root is not a file")
	}
	nodes, _ := astutil.PathEnclosingInterval(f, t.Definition.Pos(), t.Definition.End())
	return ast.Fprint(w, nil, nodes[0], nil)
}

var _ ast.Visitor = (*DefinitionFinder)(nil)

// FindDefinition finds the definition of a named variable or constant
//
// The returned struct will be nil if the definition was not found, or
// a struct locating the definition of the value otherwise
func FindDefinition(node ast.Node, name string) *DefinitionFinder {
	t := &DefinitionFinder{
		Name: name,
		Root: node,
	}

	ast.Walk(t, node)
	if t == nil {
		panic("programming error in FindDefinition")
	}
	if t.Definition == nil {
		t = nil
	}
	return t
}
