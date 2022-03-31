package pkg

import (
	"fmt"
	"go/ast"
	"strings"
)

// NeighbourVisitor ...
type NeighbourVisitor struct {
	path    string
	parser  *Parser
	todo    map[string]struct{}
	pkgName string

	importSpec map[string]*ast.ImportSpec
	locals     map[string]*Struct
}

// NewNeighbourVisitor ...
func NewNeighbourVisitor(path string, parser *Parser, todo map[string]struct{}, pkgName string) *NeighbourVisitor {
	return &NeighbourVisitor{
		path:       path,
		parser:     parser,
		todo:       todo,
		pkgName:    pkgName,
		importSpec: map[string]*ast.ImportSpec{},
		locals:     map[string]*Struct{},
	}
}

// Visit ...
func (nbv *NeighbourVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:

	case *ast.TypeSpec:
		switch specType := n.Type.(type) {
		case *ast.StructType, *ast.InterfaceType, *ast.SelectorExpr, *ast.ArrayType, *ast.MapType:
			child := NewStruct(n.Name.Name)
			child.typeSpec = n
			child.fullPath = nbv.path
			child.sourcePackageName = nbv.pkgName
			nbv.parser.addNeighbour(nbv.path, child)
			nbv.locals[n.Name.Name] = child

		case *ast.Ident:
			if n.Name.Name == "AdditionalServiceTypes" {
				fmt.Println("AdditionalServiceTypesmetttttt")
			}
			child := NewStruct(n.Name.Name)
			child.fullPath = nbv.path
			child.scalarType = n.Name.Name
			child.scalarTypeAlias = specType.Name // type MyInt int
			child.sourcePackageName = nbv.pkgName
			nbv.parser.addNeighbour(nbv.path, child)
			nbv.locals[n.Name.Name] = child
		}

	case *ast.ImportSpec:
		path := strings.Replace(n.Path.Value, "\"", "", -1)
		fields := strings.Split(path, "/")
		importName := fields[len(fields)-1]
		if n.Name != nil {
			importName = n.Name.Name
		}
		nbv.importSpec[importName] = n

		if _, isGolibrary := goLibraries[n.Path.Value]; !isGolibrary {
			nbv.todo[n.Path.Value] = struct{}{}
		}
	}
	return nbv
}
