package pkg

import (
	"fmt"
	"go/token"
	"log"
	"runtime/debug"
	"strings"

	"github.com/flosch/pongo2"
)

// render prefix: Manager, name: Manager pkg:  path:
// render prefix: ManagerpkgcmdSource, name: Source pkg: pkgcmd path: "github.com/dongzerun/cmpgen/cmd"
func (s *Struct) renderDiff(wctx *wrapperCtx, outputer *outputer) {
	if s == nil || s.visited {
		return
	}

	s.visited = true
	fmt.Printf("render struct:%s trie search:%s ignore:%v\n", s.name, wctx.fieldName, wctx.trie.Search(wctx.fieldName).IsIgnore())
	if wctx.trie.Search(wctx.fieldName).IsIgnore() {
		return
	}

	// prefix = prefix + s.importPkg + s.name
	prefix := s.importPkg + s.name
	fmt.Printf("render prefix: %s, name: %s pkg: %s members: %d path: %s\n spec: %v\n", prefix, s.name, s.importPkg, len(s.members), s.importPath, s.typeSpec)

	fn := fmt.Sprintf("Diff%s", prefix)
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
		"FuncName":  fn,
		"FieldName": s.name,
		"FieldType": fieldType,
	}

	out, err := funcTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}

	outputer.write(out)
	fmt.Printf("debuggggggstruct:%s pkg:%s path:%s fullpath:%s", s.name, s.importPkg, s.importPath, s.fullPath)
	for i := range s.membersArray {
		if !token.IsExported(s.membersArray[i]) && s.importPkg != "" {
			continue
		}

		member := s.members[s.membersArray[i]]
		fieldName := s.membersArray[i]

		body := s.renderMemberDiff(wctx.trie.Search(wctx.fieldName), outputer, fieldName, member)
		if body != "" {
			outputer.write(body)
		}
	}

	outputer.write(returnEmpty)

	if s.importPath != "" {
		outputer.appendImport(s.importPkg, s.importPath)
	}

	for i := range s.stack {
		s.stack[i].s.renderDiff(s.stack[i], outputer)
	}
}

// use reflect.DeepEqual to compare interface
func renderInterfaceDiff(outputer *outputer, fieldName string, member *Struct) string {
	outputer.appendImport("", "\"reflect\"")
	tplCtx := pongo2.Context{
		"FieldName": fieldName,
		"FieldType": member.interfaceName,
	}

	out, err := interfaceTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	return out
}

func renderFloatScalarDiff(outputer *outputer, fieldName string, member *Struct) string {
	tplCtx := pongo2.Context{
		"ItemType":  member.scalarType,
		"FieldName": fieldName,
	}

	out, err := floatCompareTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}

	funcName := "Is" + member.scalarType + "InDelta"
	if outputer.addFuncCheckIfNeedGenerated(funcName) {
		outputer.appendImport("", "\"math\"")
		outputer.writeFunc(out)
	}

	if member.pointer {
		out, err := pointerFloatTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		return out
	}

	out, err = floatCompareFuncTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	return out
}

func renderScalarDiff(outputer *outputer, fieldName string, member *Struct) string {
	tplCtx := pongo2.Context{
		"FieldName": fieldName,
	}

	if member.pointer {
		out, err := pointerTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		return out
	}

	out, err := scalarTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	return out
}

func renderArrayDiff(tr *Trie, outputer *outputer, fieldName string, member *Struct, parent *Struct) string {
	if member.scalarType == "" && member.containerItem == nil {
		log.Fatalf("invalid array type fieldname: %s name:%s updated:%v stack: %s\n", fieldName, member.name, member.updated, string(debug.Stack()))
	}

	// scalar non-pointer, []int []string []float
	if member.scalarType != "" {
		star := ""
		if member.pointer {
			star = "*"
		}

		funcName := fmt.Sprintf("diffSlice%s%s%s", member.arrayLength, star, member.scalarType)
		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		tplCtx := pongo2.Context{
			"FuncName":     funcName,
			"ItemType":     member.scalarType,
			"FullItemType": member.scalarType,
			"ArrayLength":  member.arrayLength,
			"Star":         star,
			"PkgName":      "",
		}
		out, err := sliceScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")

		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)
		return ""
	}

	if member.containerItem == nil {
		log.Fatal(fieldName + " container item empty")
	}

	item := member.containerItem

	// pointer scalar, like []*int []*string
	if item.pointer && item.scalarType != "" {
		funcName := fmt.Sprintf("diffSlice%sStar%s", member.arrayLength, item.scalarType)
		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		tplCtx := pongo2.Context{
			"ArrayLength":  member.arrayLength,
			"ItemType":     item.scalarType,
			"FullItemType": item.scalarType,
			"Star":         "*",
			"PkgName":      "",
		}
		out, err := slicePointerScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")
		return ""
	}

	// slice struct, like []struct []pkg.struct
	if item.scalarType == "" && item.pointerStruct == nil {
		structDiffFunc := fmt.Sprintf("Diff%s%s", item.importPkg, item.name)

		funcName := fmt.Sprintf("diffSlice%s%s%s", member.arrayLength, item.importPkg, item.name)
		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)
		wc := &wrapperCtx{
			s:         item,
			fieldName: fieldName,
			trie:      tr,
		}
		parent.stack = append(parent.stack, wc)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}
		star := ""
		if item.pointer {
			star = "*"
		}
		fullItemType := item.name
		if item.importPkg != "" {
			fullItemType = item.importPkg + "." + item.name
		}
		tplCtx := pongo2.Context{
			"FuncName":       funcName,
			"StructDiffFunc": structDiffFunc,
			"ArrayLength":    member.arrayLength,
			"Star":           star,
			"PkgName":        item.importPkg,
			"ItemType":       item.name,
			"FullItemType":   fullItemType,
		}
		out, err := sliceStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")
		return ""
	}

	// slice pointer struct, like []*struct []*pkg.struct
	if item.scalarType == "" && item.pointerStruct != nil {
		structDiffFunc := fmt.Sprintf("Diff%s%s", item.pointerStruct.importPkg, item.pointerStruct.name)

		funcName := fmt.Sprintf("diffSlice%sStar%s%s", member.arrayLength, item.pointerStruct.importPkg, item.pointerStruct.name)
		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)
		wc := &wrapperCtx{
			s:         item.pointerStruct,
			fieldName: fieldName,
			trie:      tr,
		}
		parent.stack = append(parent.stack, wc)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		star := ""
		if item.pointer {
			star = "*"
		}
		fullItemType := item.pointerStruct.name
		if item.pointerStruct.importPkg != "" {
			fullItemType = item.pointerStruct.importPkg + "." + item.pointerStruct.name
		}
		tplCtx := pongo2.Context{
			"FuncName":       funcName,
			"StructDiffFunc": structDiffFunc,
			"ArrayLength":    member.arrayLength,
			"Star":           star,
			"PkgName":        item.pointerStruct.importPkg,
			"ItemType":       item.pointerStruct.name,
			"FullItemType":   fullItemType,
		}
		out, err := slicePointerStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")
		return ""
	}

	return ""
}

func renderMapDiff(tr *Trie, outputer *outputer, fieldName string, member *Struct, parent *Struct) string {
	// debug map:   Photos Photos DeliveryBookingMeta
	fmt.Println("debug map: ", fieldName, member.name, parent.name, member.containerItem == nil)

	if tr.Search(fieldName).IsIgnore() {
		return ""
	}

	// map type: map[int]int map[int]string ...
	if member.scalarType != "" && member.containerItem == nil {
		keyType := member.containerKey
		valueType := member.scalarType
		funcName := fmt.Sprintf("DiffMap%s%s", strings.Replace(member.containerKey, ".", "", -1), valueType)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}
		tplCtx := pongo2.Context{
			"FuncName":  funcName,
			"KeyType":   keyType,
			"ValueType": valueType,
		}
		out, err := mapScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")
		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}

		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)
		return ""
	}

	// map type: map[string]*int ... map[string]*string
	if member.containerItem != nil && member.containerItem.pointer && member.containerItem.scalarType != "" {
		item := member.containerItem
		keyType := member.containerKey
		valueType := item.scalarType
		funcName := fmt.Sprintf("DiffMap%sStar%s", strings.Replace(member.containerKey, ".", "", -1), valueType)

		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		tplCtx := pongo2.Context{
			"FuncName":  funcName,
			"KeyType":   keyType,
			"ValueType": valueType,
			"Star":      "*",
		}
		out, err := mapPointerScalarTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")
		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}

		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)
		return ""
	}

	// map type: map[string]pkg.struct
	if member.containerItem != nil && member.containerItem.scalarType == "" && !member.containerItem.pointer {
		keyType := member.containerKey
		valueType := member.containerItem.name
		if member.containerItem.importPkg != "" {
			valueType = member.containerItem.importPkg + "." + member.containerItem.name
		}

		funcName := fmt.Sprintf("DiffMap%s%s%s", strings.Replace(member.containerKey, ".", "", -1), member.containerItem.importPkg, member.containerItem.name)
		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		structDiffFunc := fmt.Sprintf("Diff%s%s", member.containerItem.importPkg, member.containerItem.name)
		tplCtx := pongo2.Context{
			"FuncName":       funcName,
			"KeyType":        keyType,
			"ValueType":      valueType,
			"Star":           "*",
			"StructDiffFunc": structDiffFunc,
		}
		out, err := mapStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")
		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}

		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)
		outputer.appendImport(member.containerItem.importPkg, member.containerItem.importPath)
		wc := &wrapperCtx{
			s:         member.containerItem,
			fieldName: fieldName,
			trie:      tr.Search(fieldName),
		}
		parent.stack = append(parent.stack, wc)
		return ""
	}

	// map type: map[string]*pkg.struct
	if member.containerItem != nil && member.containerItem.pointerStruct != nil {
		fmt.Println("debug map container: ", member.containerItem.name, member.containerItem.pointer, member.containerItem.scalarType)
		keyType := member.containerKey
		pointer := member.containerItem.pointerStruct
		valueType := pointer.name
		if pointer.importPkg != "" {
			valueType = pointer.importPkg + "." + pointer.name
		}

		funcName := fmt.Sprintf("DiffMap%s%s%s", strings.Replace(member.containerKey, ".", "", -1), pointer.importPkg, pointer.name)
		if !outputer.addFuncCheckIfNeedGenerated(funcName) {
			return ""
		}

		structDiffFunc := fmt.Sprintf("Diff%s%s", pointer.importPkg, pointer.name)
		tplCtx := pongo2.Context{
			"FuncName":       funcName,
			"KeyType":        keyType,
			"ValueType":      valueType,
			"Star":           "*",
			"StructDiffFunc": structDiffFunc,
		}
		out, err := mapPointerStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		outputer.appendImport("", "\"unsafe\"")
		if strings.Contains(member.containerKey, ".") {
			outputer.appendImport(member.containerPkg, member.containerPath)
		}

		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, fieldName, fieldName, fieldName)
		outputer.write(text)
		outputer.appendImport(member.containerItem.pointerStruct.importPkg, member.containerItem.pointerStruct.importPath)
		wc := &wrapperCtx{
			s:         pointer,
			fieldName: fieldName,
			trie:      tr.Search(fieldName),
		}
		parent.stack = append(parent.stack, wc)
		return ""
	}
	return ""
}

func (s *Struct) renderMemberDiff(tr *Trie, outputer *outputer, m string, member *Struct) string {
	// debug prefix: fieldname: PickupPinCreatedAt membername:Time sname:DeliveryBookingMeta pkg:time path:"time"
	fmt.Printf("debug prefix fieldname: %s membername:%s sname:%s pkg:%s path:%s\n", m, member.name, s.name, member.importPkg, member.importPath)
	fmt.Printf("render member:%s trie search:%s ignore:%v\n", member.name, m, tr.Search(m).IsIgnore())
	if tr.Search(m).IsIgnore() {
		return ""
	}

	if member.interfaceName != "" {
		return renderInterfaceDiff(outputer, m, member)
	}

	// output scalar type, maybe pointer
	if (member.scalarType != "" && member.containerType == "") || (member.importPkg == "time" && member.name == "Duration") {
		if _, isFloat := floatTypes[member.scalarType]; isFloat {
			return renderFloatScalarDiff(outputer, m, member)
		}
		return renderScalarDiff(outputer, m, member)
	}

	switch member.containerType {
	case arrayType:
		return renderArrayDiff(tr, outputer, m, member, s)
	case mapType:
		return renderMapDiff(tr, outputer, m, member, s)
	}

	if member.importPkg == "time" && member.name == "Time" {
		tplCtx := pongo2.Context{
			"FieldName": m,
		}

		out, err := isTimeInDeltaTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}

		if outputer.addFuncCheckIfNeedGenerated("IsTimeInDelta") {
			outputer.writeFunc(out)
			outputer.appendImport("", "\"time\"")
		}

		out, err = diffTimeTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.write(out)
		return ""
	}

	// normal struct
	if !member.pointer && member.pointerStruct == nil {
		tplCtx := pongo2.Context{
			"Prefix":       "",
			"FullItemType": member.importPkg + member.name,
			"FieldName":    m,
		}
		out, err := diffStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}

		outputer.write(out)
		wc := &wrapperCtx{
			s:         member,
			fieldName: m,
			trie:      tr,
		}
		s.stack = append(s.stack, wc)
		return ""
	}

	// normal pointer struct, not *[]int *map[string]int
	if member.pointerStruct != nil && member.pointerStruct.containerType == "" {
		tplCtx := pongo2.Context{
			"Prefix":       "",
			"FullItemType": member.pointerStruct.importPkg + member.pointerStruct.name,
			"FieldName":    m,
		}
		out, err := diffPointerStructTemplate.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.write(out)
		wc := &wrapperCtx{
			s:         member.pointerStruct,
			fieldName: m,
			trie:      tr,
		}
		s.stack = append(s.stack, wc)
		return ""
	}

	// pointer scalar array
	if member.pointerStruct != nil &&
		member.pointerStruct.containerType == arrayType &&
		member.pointerStruct.scalarType != "" &&
		member.pointerStruct.arrayLength != "" {

		pointerStruct := member.pointerStruct

		fullItemType := pointerStruct.scalarType
		star := ""
		tpl := pointerSliceScalarTemplate

		if pointerStruct.pointer {
			star = "Star"
			fullItemType = "*" + pointerStruct.scalarType
		}
		if pointerStruct.arrayLength != "" {
			tpl = pointerArrayScalarTemplate
		}

		diffFunc := fmt.Sprintf("diffStarArray%s%s%s", pointerStruct.arrayLength, star, pointerStruct.scalarType)

		arrayCtx := pongo2.Context{
			"FieldName": m,
			"DiffFunc":  diffFunc,
		}

		out, err := diffPointerArrayTemplate.Execute(arrayCtx)
		if err != nil {
			panic(err)
		}
		outputer.write(out)

		if !outputer.addFuncCheckIfNeedGenerated(diffFunc) {
			return ""
		}
		tplCtx := pongo2.Context{
			"FieldName":    m,
			"ArrayLength":  pointerStruct.arrayLength,
			"Star":         star,
			"ItemType":     pointerStruct.scalarType,
			"FullItemType": fullItemType,
		}

		out, err = tpl.Execute(tplCtx)
		if err != nil {
			panic(err)
		}
		outputer.writeFunc(out)
		return ""
	}

	if member.pointerStruct != nil &&
		member.pointerStruct.containerType == arrayType &&
		member.pointerStruct.scalarType != "" &&
		member.pointerStruct.arrayLength == "" {

		pointerStruct := member.pointerStruct
		fmt.Printf("member.pointerStruct:%s star:%v arrl:%s\n", pointerStruct.name, pointerStruct.pointer, pointerStruct.arrayLength)
		fullItemType := pointerStruct.scalarType
		star := ""
		if pointerStruct.pointer {
			star = "Star"
			fullItemType = "*" + pointerStruct.scalarType
		}

		funcName := "diffSlice" + star + pointerStruct.scalarType
		if outputer.addFuncCheckIfNeedGenerated(funcName) {
			tplCtx := pongo2.Context{
				"Star":         star,
				"ItemType":     pointerStruct.scalarType,
				"FullItemType": fullItemType,
				"ArrayLength":  pointerStruct.arrayLength,
			}

			tpl := sliceScalarTemplate
			if pointerStruct.pointer {
				tpl = slicePointerScalarTemplate
			}
			out, err := tpl.Execute(tplCtx)
			if err != nil {
				panic(err)
			}
			outputer.writeFunc(out)
			outputer.appendImport("", "\"unsafe\"")
		}

		funcName = "diffStarSlice" + star + pointerStruct.scalarType
		if outputer.addFuncCheckIfNeedGenerated(funcName) {
			tplCtx := pongo2.Context{
				"Star":         star,
				"ItemType":     pointerStruct.scalarType,
				"FullItemType": fullItemType,
				"ArrayLength":  pointerStruct.arrayLength,
			}
			out, err := diffStarSliceScalarTemplate.Execute(tplCtx)
			if err != nil {
				panic(err)
			}
			outputer.writeFunc(out)
		}

		text := fmt.Sprintf(diffPointerUtilTemplate, funcName, m, m, m)
		outputer.write(text)
		return ""
	}

	// pointer map
	if member.pointerStruct != nil && member.pointerStruct.containerType == mapType {
	}
	return ""
}
