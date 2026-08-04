package config

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionCodeHasNoBareStringLiterals(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(fset, name, nil, 0)
		if parseErr != nil {
			t.Errorf("parse %s: %v", name, parseErr)
			continue
		}
		parents := parentNodes(file)
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING || allowedStringLiteral(literal, parents) {
				return true
			}
			position := fset.Position(literal.Pos())
			t.Errorf("bare string literal at %s: define a named constant", position)
			return true
		})
	}
}

func parentNodes(root ast.Node) map[ast.Node]ast.Node {
	parents := make(map[ast.Node]ast.Node)
	var stack []ast.Node
	ast.Inspect(root, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) > 0 {
			parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})
	return parents
}

func allowedStringLiteral(literal *ast.BasicLit, parents map[ast.Node]ast.Node) bool {
	for node := ast.Node(literal); node != nil; node = parents[node] {
		switch current := node.(type) {
		case *ast.ImportSpec:
			return true
		case *ast.Field:
			return current.Tag == literal
		case *ast.GenDecl:
			return current.Tok == token.CONST
		case *ast.FuncDecl, *ast.FuncLit:
			return false
		}
	}
	panic(fmt.Sprintf("string literal %s has no declaration parent", literal.Value))
}

func TestStringLiteralGuardCoversProductionFiles(t *testing.T) {
	if _, err := os.Stat(filepath.Join(".", "loader.go")); err != nil {
		t.Fatalf("guard must run from the config package directory: %v", err)
	}
}
