package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/dongzerun/ast-tools/pkg/config"
)

// Parser ...
type Parser struct {
	config        *config.Config
	trie          *Trie
	neighboursMap map[string]map[string]*Struct
	visitedPkg    map[string]bool

	todo map[string]struct{}

	visitedSpec map[string]*Struct
}

// NewParser ...
func NewParser(cfg *config.Config) *Parser {
	return &Parser{
		config:        cfg,
		trie:          BuildIgnoreFieldsTrie(cfg.IgnoreFields),
		neighboursMap: make(map[string]map[string]*Struct),
		visitedPkg:    make(map[string]bool),
		visitedSpec:   map[string]*Struct{},
	}
}

func filter(f fs.FileInfo) bool {
	if f.IsDir() {
		return false
	}

	fname := f.Name()
	if strings.HasSuffix(fname, "_test.go") ||
		strings.HasPrefix(fname, "mock_") ||
		strings.HasPrefix(fname, "MOCK_") {
		return false
	}
	return true
}

// Do ...
func (p *Parser) Do() error {
	switch p.config.Action {
	case "diff":
		return p.OutputDiff()
	case "copy":
		return p.OutputDeepCopy()
	case "convert":
		return p.OutputConvert()
	}
	return errors.New("didn't specifiy action")
}

// OutputDiff find target and output diff function
func (p *Parser) OutputDiff() error {
	p.IterateGenNeighbours(p.config.Dir)
	p.printNeighbour()

	s := p.IterateDir(p.config.Dir, p.config.Name)

	if s == nil {
		return errors.New("couldn't get " + p.config.Name + " struct def")
	}
	p.trie.Print(0)
	fmt.Printf("After Do, render the output\n\n")
	outputer := newOutPuter()
	abs, _ := filepath.Abs(p.config.Dir)
	outputer.targetDir = abs
	wctx := &wrapperCtx{
		s:         s,
		fieldName: p.config.Name,
		trie:      p.trie,
	}
	s.renderDiff(wctx, outputer)
	res := outputer.toString(s.sourcePackageName)
	fmt.Println(res)
	// fileName := path.Join(p.config.Dir, "z_diff_"+strings.ToLower(p.config.Name)+".go")
	// ioutil.WriteFile(fileName, []byte(res), 0644)
	return nil
}

// OutputDeepCopy find target and output deep copy function
func (p *Parser) OutputDeepCopy() error {
	p.IterateGenNeighbours(p.config.Dir)
	p.printNeighbour()

	s := p.IterateDir(p.config.Dir, p.config.Name)

	if s == nil {
		return errors.New("couldn't get " + p.config.Name + " struct def")
	}
	outputer := newOutPuter()
	abs, _ := filepath.Abs(p.config.Dir)
	outputer.targetDir = abs

	s.renderCopy(outputer)
	res := outputer.toString(s.sourcePackageName)
	fmt.Println(res)
	return nil
}

// OutputConvert find target and output convert function
func (p *Parser) OutputConvert() error {
	if p.config.SrcPkg == "" ||
		p.config.Source == "" ||
		p.config.TargetPkg == "" ||
		p.config.Target == "" {
		return errors.New("convert params empty")
	}

	if !filepath.IsAbs(p.config.SrcPkg) {
		p.config.SrcPkg = path.Join(os.Getenv("GOPATH"), "src", p.config.SrcPkg)
	}

	if !filepath.IsAbs(p.config.TargetPkg) {
		p.config.TargetPkg = path.Join(os.Getenv("GOPATH"), "src", p.config.TargetPkg)
	}

	p.IterateGenNeighbours(p.config.SrcPkg)
	p.IterateGenNeighbours(p.config.TargetPkg)

	s := p.IterateDir(p.config.SrcPkg, p.config.Source)

	if s == nil {
		return errors.New("couldn't get " + p.config.SrcPkg + "." + p.config.Source)
	}

	t := p.IterateDir(p.config.TargetPkg, p.config.Target)

	if t == nil {
		return errors.New("couldn't get " + p.config.TargetPkg + "." + p.config.Target)
	}

	outputer := newOutPuter()
	c := newConvert()
	c.renderConvert(s, t, p.config.SrcTag, p.config.TargetTag, outputer)
	outputdir, err := filepath.Abs(p.config.Output)
	if err != nil {
		outputdir = path.Join(os.Getenv("GOPATH"), "src", p.config.Output)
	}
	res := outputer.toString(outputdir)
	fmt.Println(res)

	fileName := path.Join(outputdir, "z_"+strings.ToLower(p.config.Source)+".go")
	ioutil.WriteFile(fileName, []byte(res), 0644)
	// cmd := exec.Command("gofmt", "-w", "-s", fileName)
	// _, err = cmd.CombinedOutput()

	buf := &bytes.Buffer{}
	c.report(buf, 0)
	fmt.Println(buf.String())
	return err
}

func (p *Parser) addNeighbour(path string, neighbour *Struct) {
	// fmt.Printf("addNeighbour path: %s neighbour:%s\n", path, neighbour.name)
	_, exists := p.neighboursMap[path]
	if !exists {
		p.neighboursMap[path] = make(map[string]*Struct)
	}
	p.neighboursMap[path][neighbour.name] = neighbour
}

func (p *Parser) getNeighbour(path, name string) *Struct {
	fmt.Println("getNeighbour start search ", path, name)
	if _, visited := p.visitedSpec[path+name]; visited {
		return p.visitedSpec[path+name]
	}

	_, exists := p.neighboursMap[path]
	if !exists {
		log.Fatalf("get neighbour path %s name %s not exists failed\n %s", path, name, string(debug.Stack()))
	}
	n, exists := p.neighboursMap[path][name]
	if !exists {
		log.Fatalf("get neighbour path %s name %s not exists failed\n %s", path, name, string(debug.Stack()))
	}
	if !n.updated {
		p.updateTarget(n)
	}

	p.visitedSpec[path+name] = n
	return n
}

// printNeighbour path:/Users/zerun.dong/gopath/src/github.com/json-iterator/go name:trueAny
func (p *Parser) printNeighbour() {
	// for path := range p.neighboursMap {
	// 	for name := range p.neighboursMap[path] {
	// 		fmt.Printf("printNeighbour path:%s name:%s\n", path, name)
	// 	}
	// }
}

// IterateGenNeighbours ...
func (p *Parser) IterateGenNeighbours(dir string) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return
	}

	p.visitedPkg[dir] = true

	pkgs, err := parser.ParseDir(token.NewFileSet(), path, filter, 0)
	if err != nil {
		return
	}

	todo := map[string]struct{}{}
	for pkgName, pkg := range pkgs {
		nbv := NewNeighbourVisitor(path, p, todo, pkgName)
		for _, astFile := range pkg.Files {
			ast.Walk(nbv, astFile)
		}

		// update import specs per file
		for name := range nbv.locals {
			fmt.Sprintf("IterateGenNeighbours find struct:%s pkg:%s path:%s\n", name, nbv.locals[name].importPkg, nbv.locals[name].importPath)
			nbv.locals[name].importSpecs = nbv.importSpec
		}
	}

	for path := range todo {
		dir := os.Getenv("GOPATH") + "/src/" + strings.Replace(path, "\"", "", -1)
		if _, visited := p.visitedPkg[dir]; visited {
			continue
		}
		p.IterateGenNeighbours(dir)
	}
}

// IterateDir dir to get target typeSpec
func (p *Parser) IterateDir(dir string, targetName string) *Struct {
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}

	return p.getNeighbour(path, targetName)
}

func (p *Parser) updateTarget(target *Struct) {
	// if it's scalar type, just return
	if target.scalarType != "" || target.updated {
		return
	}

	target.updated = true
	fmt.Printf("debug updateTarget struct: %s  path: %s\n", target.name, target.fullPath)

	if target.typeSpec != nil {
		switch specType := target.typeSpec.Type.(type) {
		case *ast.StructType:
			// parse Fields if it's struct type
			for _, field := range specType.Fields.List {
				if len(field.Names) > 0 {
					if !token.IsExported(field.Names[0].Name) {
						continue
					}
				}

				switch ft := field.Type.(type) {
				case *ast.Ident:
					p.parseScalar(field, ft, target)

				case *ast.SelectorExpr:
					p.parseStruct(field, ft, target, nil)

				case *ast.StarExpr:
					p.parsePointer(field, ft, target, nil)

				case *ast.ArrayType:
					p.parseArray(field, ft, target, nil)

				case *ast.MapType:
					p.parseMap(field, ft, target, nil)
				}
			}

		case *ast.InterfaceType:
			target.name = target.typeSpec.Name.Name
			target.interfaceName = target.typeSpec.Name.Name

		// type Lists []*List
		case *ast.ArrayType:

		// type Maps []*Map
		case *ast.MapType:
		}
	}

	if target.pointerStruct != nil {
		p.updateTarget(target.pointerStruct)
	}

	if target.containerItem != nil {
		p.updateTarget(target.containerItem)
	}
}

func (p *Parser) getImportNameAndPath(fullpath string) (string, string) {
	targetPath := p.config.Dir
	absPath, _ := filepath.Abs(targetPath)
	if fullpath == absPath {
		return "", ""
	}

	pkg := filepath.Base(fullpath)
	prefix := os.Getenv("GOPATH") + "/src/"
	if len(prefix) < len(fullpath) {
		return pkg, "\"" + fullpath[len(prefix):] + "\""
	}
	return pkg, ""
}

// mt *ast.MapType indicates it's a map type
func (p *Parser) parseMap(field *ast.Field, mt *ast.MapType, parent *Struct, pointer *Struct) {
	if (parent == nil && pointer == nil) || (parent != nil && pointer != nil) {
		log.Fatal("parent and pointer should one nil, another not nil ")
	}

	fieldName := field.Names[0].Name
	member := NewStruct(fieldName)
	member.containerType = mapType
	member.tag = getTagFromASTField(field)
	if parent != nil {
		member.fullPath = parent.fullPath
		member.importSpecs = parent.importSpecs
	} else {
		member.fullPath = pointer.fullPath
		member.importSpecs = pointer.importSpecs
	}

	importStruct := parent
	if parent == nil {
		importStruct = pointer
	}

	switch key := mt.Key.(type) {
	case *ast.SelectorExpr:
		importPkg := key.X.(*ast.Ident).Name
		importSpec, exists := importStruct.importSpecs[importPkg]
		if !exists {
			log.Fatal("import "+importPkg+" not exists", string(debug.Stack()))
		}

		// realPkg := getImportPkgFromSelectorSel(importSpec.Path.Value)
		member.containerKey = key.X.(*ast.Ident).Name + "." + key.Sel.Name
		member.containerPkg = importPkg
		member.containerPath = importSpec.Path.Value

	case *ast.Ident:
		member.containerKey = key.Name
	}

	if parent != nil {
		parent.addMember(fieldName, member)
	}

	if pointer != nil {
		pointer.pointerStruct = member
	}

	switch value := mt.Value.(type) {
	// map[string]A.B
	case *ast.SelectorExpr:
		p.parseStruct(field, value, nil, member)

	// map[string]*A.B or map[string]*int
	case *ast.StarExpr:
		p.parsePointer(field, value, nil, member)

	// map[string]int
	case *ast.Ident:
		if _, isScalar := scalarTypes[value.Name]; isScalar {
			member.scalarType = value.Name
		} else {
			n := p.getNeighbour(importStruct.fullPath, value.Name)
			fmt.Println("parseScalar start get neighbour ", n.name)
			n.fullPath = member.fullPath
			n.importSpecs = member.importSpecs
			// n.importPkg = importStruct.importPkg
			// n.importPath = importStruct.importPath
			n.importPkg, n.importPath = p.getImportNameAndPath(importStruct.fullPath)
			member.containerItem = n
		}
	}
}

// *ast.Ident indicates it's a scalar type, or struct in same package
func (p *Parser) parseScalar(field *ast.Field, ident *ast.Ident, parent *Struct) {
	fieldName := ident.Name
	if len(field.Names) > 0 {
		// it's not anonymous field
		fieldName = field.Names[0].Name
	}

	if fieldName == "AdditionalServiceTypes" {
		fmt.Println("parseScalar start ", fieldName, ident.Name)
	}

	tag := getTagFromASTField(field)

	if _, scalar := scalarTypes[ident.Name]; scalar {
		member := NewStruct(fieldName)
		if ident.Name == "error" {
			member.interfaceName = "error"
		} else {
			member.scalarType = ident.Name
		}

		member.tag = tag
		parent.addMember(fieldName, member)
	} else {
		n := p.getNeighbour(parent.fullPath, ident.Name)
		n.tag = tag
		n.fullPath = parent.fullPath
		// n.importPkg = parent.importPkg
		// n.importPath = parent.importPath
		n.importPkg, n.importPath = p.getImportNameAndPath(parent.fullPath)

		fmt.Printf("parseScalar name:%s fullpath:%s pname:%s importpkg:%s ppkg:%s importpath:%s ppath:%s \n", n.name, n.fullPath, parent.name, n.importPkg, parent.importPkg, n.importPath, parent.importPath)
		parent.addMember(fieldName, n)
	}
}

// *ast.StarExpr means *int, *struct, *[]int, *map[int]ing ....
func (p *Parser) parsePointer(field *ast.Field, star *ast.StarExpr, parent *Struct, pointer *Struct) {
	if (parent == nil && pointer == nil) || (parent != nil && pointer != nil) {
		log.Fatal("parent and pointer should one nil, another not nil ")
	}

	fieldName := ""
	if len(field.Names) > 0 {
		fieldName = field.Names[0].Name
	}

	neighbourStruct := parent
	if parent == nil {
		neighbourStruct = pointer
	}

	// currently fieldName maybe empty, so need fill it later
	member := NewStruct(fieldName)
	member.pointer = true
	member.tag = getTagFromASTField(field)
	member.fullPath = neighbourStruct.fullPath

	switch starX := star.X.(type) {
	case *ast.Ident:
		if fieldName == "" {
			fieldName = starX.Name
		}

		if _, scalar := scalarTypes[starX.Name]; scalar {
			member.scalarType = starX.Name
		} else {
			n := p.getNeighbour(neighbourStruct.fullPath, starX.Name)
			n.fullPath = neighbourStruct.fullPath
			// n.importPkg = neighbourStruct.importPkg
			// n.importPath = neighbourStruct.importPath
			n.importPkg, n.importPath = p.getImportNameAndPath(neighbourStruct.fullPath)
			member.pointerStruct = n
		}

	case *ast.SelectorExpr:
		if fieldName == "" {
			fieldName = starX.Sel.Name
		}

		importPkg := starX.X.(*ast.Ident).Name
		importSpec, exists := neighbourStruct.importSpecs[importPkg]
		if !exists {
			log.Fatal("import "+importPkg+" not exists", string(debug.Stack()))
		}

		var selectorTarget *Struct
		typeName := starX.Sel.Name
		realPkg := getImportPkgFromSelectorSel(importSpec.Path.Value)
		_, isGoBuiltinType := goLibraryTypes[realPkg+"."+starX.Sel.Name]
		if isGoBuiltinType {
			selectorTarget = NewStruct(starX.Sel.Name)
		} else {
			dir := getDirFromImportSpec(importSpec.Path.Value)
			selectorTarget = p.IterateDir(dir, typeName)
		}

		if selectorTarget == nil {
			fmt.Println("bug selectorTarget nil")
			break
		}

		selectorTarget.importPath = importSpec.Path.Value
		selectorTarget.importPkg = importPkg
		selectorTarget.fullPath = getDirFromImportSpec(importSpec.Path.Value)
		member.pointerStruct = selectorTarget

	case *ast.ArrayType:
		p.parseArray(field, starX, nil, member)

	case *ast.MapType:
		p.parseMap(field, starX, nil, member)
	}

	if parent != nil {
		parent.addMember(fieldName, member)
	}

	if pointer != nil {
		if pointer.containerType != "" {
			pointer.containerItem = member
		}

		if pointer.pointer {
			pointer.pointerStruct = member
		}
	}
}

// ast.ArrayType means it's a array, []int []string []pkg.struct []*int ...
func (p *Parser) parseArray(field *ast.Field, array *ast.ArrayType, parent *Struct, pointer *Struct) {
	if (parent == nil && pointer == nil) || (parent != nil && pointer != nil) {
		log.Fatal("parent and pointer should one nil, another not nil ")
	}

	neighbourStruct := parent
	if parent == nil {
		neighbourStruct = pointer
	}

	fieldName := field.Names[0].Name
	member := NewStruct(fieldName)

	member.containerType = "ArrayType"
	member.tag = getTagFromASTField(field)
	member.fullPath = neighbourStruct.fullPath

	if parent != nil {
		member.importSpecs = parent.importSpecs
	} else {
		member.importSpecs = pointer.importSpecs
	}

	// means it's a array not slice
	if array.Len != nil {
		member.arrayLength = array.Len.(*ast.BasicLit).Value
	}

	if parent != nil {
		parent.addMember(fieldName, member)
	}

	if pointer != nil {
		pointer.pointerStruct = member
	}

	switch elt := array.Elt.(type) {
	// []*pkg.struct
	case *ast.StarExpr:
		p.parsePointer(field, elt, nil, member)

	// []pkg.struct
	case *ast.SelectorExpr:
		p.parseStruct(field, elt, nil, member)

	// todo: 还得写一个本 pkg 内的值类型
	// array 普通类型
	// []int []string etc...
	case *ast.Ident:
		if _, scalar := scalarTypes[elt.Name]; scalar {
			if elt.Name == "error" {
				member.interfaceName = "error"
			} else {
				member.scalarType = elt.Name
			}
		} else {
			n := p.getNeighbour(neighbourStruct.fullPath, elt.Name)
			n.importPkg, n.importPath = p.getImportNameAndPath(neighbourStruct.fullPath)
			if parent != nil {
				n.fullPath = parent.fullPath
				member.containerItem = n
			}

			if pointer != nil {
				n.fullPath = pointer.fullPath
				// pointer.pointerStruct = n
			}
		}
	}
}

// *ast.SelectorExpr means it's A.B, A is pkg name, B is type
func (p *Parser) parseStruct(field *ast.Field, selector *ast.SelectorExpr, parent *Struct, pointer *Struct) {
	if (parent == nil && pointer == nil) || (parent != nil && pointer != nil) {
		log.Fatal("parent and pointer should one nil, another not nil ")
	}

	importPkgName := selector.X.(*ast.Ident).Name
	fieldName := selector.Sel.Name
	if len(field.Names) > 0 {
		fieldName = field.Names[0].Name
	}

	importSpace := parent
	fmt.Println("import space is parent ", fieldName)
	if importSpace == nil {
		fmt.Println("import space is pointer ", fieldName)
		importSpace = pointer
	}

	importSpec, exists := importSpace.importSpecs[importPkgName]
	if !exists {
		log.Fatal("import "+importPkgName+" not exists", string(debug.Stack()))
	}

	realPkg := getImportPkgFromSelectorSel(importSpec.Path.Value)

	var member *Struct
	_, isGoBuiltinType := goLibraryTypes[realPkg+"."+selector.Sel.Name]

	if isGoBuiltinType {
		member = NewStruct(selector.Sel.Name)
	} else {
		dir := getDirFromImportSpec(importSpec.Path.Value)
		// member should not nil
		member = p.IterateDir(dir, selector.Sel.Name)
	}

	if member == nil {
		fmt.Println("bug member is nil")
		return
	}
	member.importPkg = importPkgName
	member.fullPath = getDirFromImportSpec(importSpec.Path.Value)
	// this means import has alias, like:
	// aliaspkg "github.com/dongzerun/cmpgen/pkg"
	if !strings.HasSuffix(importSpec.Path.Value, "/"+importPkgName+"\"") {
		member.importAlias = importPkgName
	}
	member.importPath = importSpec.Path.Value
	member.tag = getTagFromASTField(field)

	if parent != nil {
		parent.addMember(fieldName, member)
	}

	if pointer != nil {
		switch pointer.containerType {
		case "ArrayType", "MapType":
			pointer.containerItem = member
		default:
			pointer.pointerStruct = member
		}
	}
}
