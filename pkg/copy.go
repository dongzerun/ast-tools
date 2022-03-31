package pkg

import (
	"fmt"
	"log"
	"strings"

	"go/token"

	"github.com/flosch/pongo2"
)

func (s *Struct) renderCopy(outputer *outputer) {
	if s == nil || s.visited {
		return
	}

	s.visited = true
	fmt.Printf("start renderCopy struct: %s\n", s.name)
	fn := fmt.Sprintf("Copy%s%s", s.importPkg, s.name)
	if !outputer.addFuncCheckIfNeedGenerated(fn) {
		return
	}

	// the is struct entry, print the func signature
	fieldType := s.name
	if s.importAlias != "" {
		fieldType = s.importAlias
	}

	if s.importPkg != "" {
		fieldType = s.importPkg + "." + s.name
	}

	if outputer.targetDir != s.fullPath {
		fieldType = s.importPkg + "." + s.name
		outputer.appendImport(s.importPkg, s.importPath)
	}

	tplCtx := pongo2.Context{
		"PkgName":      s.importPkg,
		"ItemType":     s.name,
		"FullItemType": fieldType,
	}

	out, err := copyFuncTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}

	outputer.write(out)
	for i := range s.membersArray {
		if !token.IsExported(s.membersArray[i]) && s.importPkg != "" {
			continue
		}

		member := s.members[s.membersArray[i]]
		fieldName := s.membersArray[i]
		fmt.Printf("iter members:%s structname:%s\n", fieldName, member.name)
		body := s.renderMemberCopy(outputer, fieldName, member)
		if body != "" {
			outputer.write(body)
		}
	}

	outputer.write(copyFuncReturnString)

	if s.importPath != "" {
		outputer.appendImport(s.importPkg, s.importPath)
	}

	for i := range s.stack {
		s.stack[i].s.renderCopy(outputer)
	}
}

func (s *Struct) renderInterfaceCopy(outputer *outputer, fieldName string, member *Struct) string {
	outputer.appendImport("", "\"github.com/mohae/deepcopy\"")
	fullItemType := member.interfaceName
	if member.importPkg != "" {
		fullItemType = member.importPkg + "." + member.interfaceName
		outputer.appendImport("", member.importPath)
	}

	tplCtx := pongo2.Context{
		"FieldName":    fieldName,
		"FullItemType": fullItemType,
	}

	out, err := copyInterfaceTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	return out
}

func (s *Struct) renderScalarCopy(outputer *outputer, fieldName string, member *Struct) string {
	tplCtx := pongo2.Context{
		"FieldName": fieldName,
	}

	if member.pointer {
		out, err := copyPointerScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		return out
	}

	out, err := copyScalarTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	return out
}

func (s *Struct) renderStructCopy(outputer *outputer, fieldName string, member *Struct) string {
	tplCtx := pongo2.Context{
		"FieldName": fieldName,
		"PkgName":   member.importPkg,
		"ItemType":  member.name,
	}
	out, err := copyStructTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)

	wc := &wrapperCtx{
		s: member,
	}
	s.stack = append(s.stack, wc)
	return ""
}

func (s *Struct) renderPointerStructCopy(outputer *outputer, fieldName string, member *Struct) string {
	tplCtx := pongo2.Context{
		"FieldName": fieldName,
		"PkgName":   member.importPkg,
		"ItemType":  member.name,
	}
	out, err := copyPointerStructTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)

	wc := &wrapperCtx{
		s: member,
	}
	s.stack = append(s.stack, wc)
	return ""
}

func renderMapCopy(outputer *outputer, fieldName string, member *Struct, parent *Struct) string {
	// map type: map[int]int map[int]string ...
	if member.scalarType != "" && member.containerItem == nil {
		keyType := member.containerKey
		valueType := member.scalarType
		funcName := fmt.Sprintf("CopyMap%s%s", strings.Replace(member.containerKey, ".", "", -1), valueType)

		text, err := copyMapFieldTemplate.Execute(pongo2.Context{
			"FuncName":  funcName,
			"FieldName": fieldName,
		})
		if err != nil {
			panic(err)
		}
		outputer.write(text)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}
		tplCtx := pongo2.Context{
			"FuncName":  funcName,
			"KeyType":   keyType,
			"ValueType": valueType,
		}

		out, err := copyMapScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)

		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}
		return ""
	}

	// map type: map[string]*int ... map[string]*string
	if member.containerItem != nil && member.containerItem.pointer && member.containerItem.scalarType != "" {
		item := member.containerItem
		keyType := member.containerKey
		valueType := item.scalarType
		funcName := fmt.Sprintf("CopyMap%sStar%s", strings.Replace(member.containerKey, ".", "", -1), valueType)

		text, err := copyMapFieldTemplate.Execute(pongo2.Context{
			"FuncName":  funcName,
			"FieldName": fieldName,
		})
		if err != nil {
			panic(err)
		}
		outputer.write(text)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		tplCtx := pongo2.Context{
			"FuncName":  funcName,
			"KeyType":   keyType,
			"ValueType": valueType,
		}
		out, err := copyMapStarScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}
		return ""
	}

	// map type: map[string]pkg.struct
	if member.containerItem != nil && member.containerItem.scalarType == "" && !member.containerItem.pointer {
		keyType := member.containerKey
		valueType := member.containerItem.name
		if member.containerItem.importPkg != "" {
			valueType = member.containerItem.importPkg + "." + member.containerItem.name
		}

		funcName := fmt.Sprintf("CopyMap%s%s%s", strings.Replace(member.containerKey, ".", "", -1), member.containerItem.importPkg, member.containerItem.name)

		text, err := copyMapFieldTemplate.Execute(pongo2.Context{
			"FuncName":  funcName,
			"FieldName": fieldName,
		})
		if err != nil {
			panic(err)
		}
		outputer.write(text)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		structCopyFunc := fmt.Sprintf("Copy%s%s", member.containerItem.importPkg, member.containerItem.name)
		tplCtx := pongo2.Context{
			"FuncName":       funcName,
			"KeyType":        keyType,
			"ValueType":      valueType,
			"CopyStructFunc": structCopyFunc,
		}

		out, err := copyMapStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)

		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}

		wc := &wrapperCtx{
			s: member.containerItem,
		}
		parent.stack = append(parent.stack, wc)
		return ""
	}

	// map type: map[string]*pkg.struct
	if member.containerItem != nil && member.containerItem.pointerStruct != nil {
		keyType := member.containerKey
		pointer := member.containerItem.pointerStruct
		valueType := pointer.name
		if pointer.importPkg != "" {
			valueType = pointer.importPkg + "." + pointer.name
		}

		funcName := fmt.Sprintf("CopyMap%sStar%s%s", strings.Replace(member.containerKey, ".", "", -1), pointer.importPkg, pointer.name)
		text, err := copyMapFieldTemplate.Execute(pongo2.Context{
			"FuncName":  funcName,
			"FieldName": fieldName,
		})
		if err != nil {
			panic(err)
		}
		outputer.write(text)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		CopyStructFunc := fmt.Sprintf("Copy%s%s", pointer.importPkg, pointer.name)
		tplCtx := pongo2.Context{
			"FuncName":       funcName,
			"KeyType":        keyType,
			"ValueType":      valueType,
			"CopyStructFunc": CopyStructFunc,
		}
		out, err := copyMapStarStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}

		wc := &wrapperCtx{
			s: pointer,
		}
		parent.stack = append(parent.stack, wc)
		return ""
	}
	return ""
}

func renderArrayCopy(outputer *outputer, fieldName string, member *Struct, parent *Struct) string {
	if member.scalarType != "" {
		tplCtx := pongo2.Context{
			"FieldName":   fieldName,
			"PkgName":     member.importPkg,
			"ItemType":    member.scalarType,
			"ArrayLength": member.arrayLength,
		}
		out, err := copySliceScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.write(out)
		return ""
	}

	if member.containerItem == nil {
		log.Fatal(fieldName + " container item empty")
	}

	item := member.containerItem

	// pointer scalar, like []*int []*string
	if item.pointer && item.scalarType != "" {
		tplCtx := pongo2.Context{
			"FieldName":   fieldName,
			"PkgName":     item.importPkg,
			"ItemType":    item.scalarType,
			"ArrayLength": item.arrayLength,
		}
		out, err := copySliceStarScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.write(out)
		return ""
	}

	// slice struct, like []struct []pkg.struct
	if item.scalarType == "" && item.pointerStruct == nil {
		structCopyFunc := fmt.Sprintf("Copy%s%s", item.importPkg, item.name)

		star := ""
		if item.pointer {
			star = "*"
		}
		fullItemType := item.name
		if item.importPkg != "" {
			fullItemType = item.importPkg + "." + item.name
		}
		tplCtx := pongo2.Context{
			"StructCopyFunc": structCopyFunc,
			"ArrayLength":    member.arrayLength,
			"Star":           star,
			"PkgName":        item.importPkg,
			"ItemType":       item.name,
			"FullItemType":   fullItemType,
			"FieldName":      fieldName,
		}
		out, err := copySliceStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.write(out)

		wc := &wrapperCtx{
			s: item,
		}
		parent.stack = append(parent.stack, wc)
		return ""
	}

	// slice pointer struct, like []*struct []*pkg.struct
	if item.scalarType == "" && item.pointerStruct != nil {
		structCopyFunc := fmt.Sprintf("Copy%s%s", item.pointerStruct.importPkg, item.pointerStruct.name)

		star := ""
		if item.pointerStruct.pointer {
			star = "*"
		}
		fullItemType := item.pointerStruct.name
		if item.pointerStruct.importPkg != "" {
			fullItemType = item.pointerStruct.importPkg + "." + item.pointerStruct.name
		}
		tplCtx := pongo2.Context{
			"StructCopyFunc": structCopyFunc,
			"ArrayLength":    member.arrayLength,
			"Star":           star,
			"PkgName":        item.pointerStruct.importPkg,
			"ItemType":       item.pointerStruct.name,
			"FullItemType":   fullItemType,
			"FieldName":      fieldName,
		}
		out, err := copySliceStarStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.write(out)

		wc := &wrapperCtx{
			s: item.pointerStruct,
		}
		parent.stack = append(parent.stack, wc)
		return ""
	}

	return ""
}

func (s *Struct) renderMemberCopy(outputer *outputer, fieldName string, member *Struct) string {
	if member.interfaceName != "" {
		return s.renderInterfaceCopy(outputer, fieldName, member)
	}

	// output scalar type, maybe pointer
	if member.scalarType != "" && member.containerType == "" {
		return s.renderScalarCopy(outputer, fieldName, member)
	}

	// normal struct
	if !member.pointer && member.pointerStruct == nil && member.containerType == "" {
		return s.renderStructCopy(outputer, fieldName, member)
	}

	// normal pointer struct, not *[]int *map[string]int
	if member.pointerStruct != nil && member.pointerStruct.containerType == "" {
		return s.renderPointerStructCopy(outputer, fieldName, member.pointerStruct)
	}
	if member.containerType != "" {
		fmt.Println(member.containerType, member.containerItem)
	}

	switch member.containerType {
	case arrayType:
		return renderArrayCopy(outputer, fieldName, member, s)
	case mapType:
		return renderMapCopy(outputer, fieldName, member, s)
	}

	return ""
}
