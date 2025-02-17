package ifacecapture

import (
	"fmt"
	"go/ast"
)

// TypeChain traverses a chain of selector expressions like mypkg.MyInterface
// and collects the idents of the types in the chain.
type TypeChain struct {
	Types []*ast.Ident
}

func NewTypeChain() TypeChain {
	return TypeChain{
		Types: []*ast.Ident{},
	}
}

func (t *TypeChain) ProcessTypeChain(expr ast.Expr) error {
	switch expr.(type) {
	case *ast.Ident:
		t.Types = append(t.Types, expr.(*ast.Ident))
	case *ast.SelectorExpr:
		selExpr := expr.(*ast.SelectorExpr)
		idents := []*ast.Ident{}
		err := traverseSelChain(&idents, selExpr)
		if err != nil {
			return err
		}

		// reverse the order of the idents
		for i, j := 0, len(idents)-1; i < j; i, j = i+1, j-1 {
			idents[i], idents[j] = idents[j], idents[i]
		}

		t.Types = append(t.Types, idents...)
	}

	return nil
}

func traverseSelChain(idents *[]*ast.Ident, selExpr *ast.SelectorExpr) error {
	*idents = append(*idents, selExpr.Sel)
	switch selExpr.X.(type) {
	case *ast.Ident:
		*idents = append(*idents, selExpr.X.(*ast.Ident))
		return nil
	case *ast.SelectorExpr:
		return traverseSelChain(idents, selExpr.X.(*ast.SelectorExpr))
	default:
		return fmt.Errorf("Expected identifier, got %T", selExpr.X)
	}
}

// True if the last element of the chain is an interface.
func (t TypeChain) IsInterface() bool {
	// last in the chain is an interface
	last := t.Last()
	if last == nil {
		return false
	}

	obj := last.Obj
	if obj == nil {
		return false
	}
	typeSpec, ok := obj.Decl.(*ast.TypeSpec)
	if !ok {
		return false
	}

	_, ok = typeSpec.Type.(*ast.InterfaceType)
	return ok
}

func (t TypeChain) Last() *ast.Ident {
	if len(t.Types) == 0 {
		return nil
	}

	return t.Types[len(t.Types)-1]
}
