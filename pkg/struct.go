package pkg

import (
	"fmt"
	"go/ast"
	"log"
	"reflect"
	"runtime/debug"
	"strings"
)

const (
	mapType   = "MapType"
	arrayType = "ArrayType"
)

// Struct ...
type Struct struct {
	name string
	// represents tag like `mapstructure:"tag",json:"Tag"`
	tag reflect.StructTag
	// represents for the struct or interface def
	typeSpec *ast.TypeSpec
	// if it's a interface, not empty
	interfaceName string

	pointer       bool
	pointerStruct *Struct

	// type XXXX A.B
	aliasStruct *Struct

	// slice or map related
	containerType string  //ArrayType or MapType
	containerKey  string  // if map this is key type
	containerPkg  string  // used for map key
	containerPath string  // used for map key
	containerItem *Struct // slice&&array&map value type
	arrayLength   string  // arrayLength -1 indicates it's a arry not slice

	// string is struct field name, membersArray used to output in defined order
	members          map[string]*Struct
	membersArray     []string
	membersByTags    []*memberWithFieldName
	membersByTagsMap map[string]*memberWithFieldName

	// neighbours
	neighbours map[string]*Struct
	functions  map[string]*ast.FuncDecl

	// string is import alias name or the path name
	importSpecs map[string]*ast.ImportSpec

	// if it's a scalar type, scalarType not empty: int string bool etc...
	scalarType      string
	scalarTypeAlias string // like, type Source int, Source is just the alias of int

	importPkg         string
	importAlias       string
	importPath        string
	fullPath          string
	sourcePackageName string

	// used to output generate codes, bfs iterate fields
	stack   []*wrapperCtx
	visited bool
	updated bool
	err     error
}

type memberWithFieldName struct {
	tagName   string
	fieldName string
	member    *Struct
}

type wrapperCtx struct {
	s         *Struct
	t         *Struct
	fieldName string
	trie      *Trie
}

// NewStruct ...
func NewStruct(name string) *Struct {
	return &Struct{
		name:             name,
		members:          make(map[string]*Struct),
		membersArray:     []string{},
		neighbours:       make(map[string]*Struct),
		importSpecs:      make(map[string]*ast.ImportSpec),
		functions:        make(map[string]*ast.FuncDecl),
		stack:            []*wrapperCtx{},
		membersByTags:    []*memberWithFieldName{},
		membersByTagsMap: map[string]*memberWithFieldName{},
	}
}

func (s *Struct) updateNeighbours() {
	for _, n := range s.neighbours {
		if s.name == n.name {
			continue
		}
		n.importSpecs = s.importSpecs
		n.importPkg = s.importPkg
		n.importPath = s.importPath
		n.fullPath = s.fullPath
		n.neighbours = s.neighbours
	}
}

func (s *Struct) addMember(name string, member *Struct) {
	if _, exists := s.members[name]; exists {
		log.Fatal("already add members "+name, string(debug.Stack()))
	}
	s.members[name] = member
	s.membersArray = append(s.membersArray, name)
}

func (s *Struct) buildImportPath() {
	if s.importPkg != "" && s.importPath != "" {
		return
	}

	s.importPkg, s.importPath = getImportFromFullPath(s.fullPath)

	if s.pointerStruct != nil {
		s.pointerStruct.buildImportPath()
	}

	if s.containerItem != nil {
		s.containerItem.buildImportPath()
	}
}

func (s *Struct) buildTag(tag string) {
	fmt.Printf("start to buildTag %s\n", s.name)
	scalars := []*memberWithFieldName{}
	structs := []*memberWithFieldName{}
	pointerStructs := []*memberWithFieldName{}
	arrays := []*memberWithFieldName{}
	mapss := []*memberWithFieldName{}
	others := []*memberWithFieldName{}

	for idx := range s.membersArray {
		name := s.membersArray[idx]
		value, ok := s.members[name].tag.Lookup(tag)

		if !ok || value == "-" {
			continue
		}
		fmt.Printf("buildTag name:%s tag:%s value:%s\n", name, tag, value)
		fields := strings.Split(value, ",")
		mwf := &memberWithFieldName{tagName: fields[0], fieldName: name, member: s.members[name]}
		s.membersByTagsMap[fields[0]] = mwf

		switch {
		case s.members[name].isScalar():
			scalars = append(scalars, mwf)
		case s.members[name].isStruct() || s.members[name].isStarScalar():
			structs = append(structs, mwf)
		case s.members[name].isStarStruct():
			pointerStructs = append(pointerStructs, mwf)
		case s.members[name].isArray():
			arrays = append(arrays, mwf)
		case s.members[name].isMap():
			mapss = append(mapss, mwf)
		default:
			others = append(others, mwf)
		}
	}
	s.membersByTags = append(s.membersByTags, scalars...)
	s.membersByTags = append(s.membersByTags, pointerStructs...)
	s.membersByTags = append(s.membersByTags, structs...)
	s.membersByTags = append(s.membersByTags, arrays...)
	s.membersByTags = append(s.membersByTags, mapss...)
	s.membersByTags = append(s.membersByTags, others...)
}

func (s *Struct) isScalar() bool {
	return s.scalarType != "" && s.containerType == "" && s.pointerStruct == nil
}

func (s *Struct) isStarScalar() bool {
	return s.scalarType != "" && s.containerType == "" && s.pointerStruct == nil && s.pointer
}

func (s *Struct) isStruct() bool {
	return !s.pointer && s.pointerStruct == nil && s.containerType == ""
}

func (s *Struct) isStarStruct() bool {
	return s.pointerStruct != nil && s.pointerStruct.containerType == ""
}

func (s *Struct) isArray() bool {
	return s.containerType == arrayType
}

func (s *Struct) isMap() bool {
	return s.containerType == mapType
}
