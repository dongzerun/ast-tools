package pkg

import (
	"fmt"
	"strings"

	"go/token"

	"github.com/flosch/pongo2"
)

type convert struct {
	structName string
	reports    []*report
	reported   map[string]struct{}
	children   map[string]*convert
	stack      []*wrapperCtx
}

func newConvert() *convert {
	return &convert{
		reported: map[string]struct{}{},
		stack:    []*wrapperCtx{},
		reports:  []*report{},
		children: map[string]*convert{},
	}
}

func (c *convert) renderConvert(src, target *Struct, srcTag, targetTag string, outputer *outputer) {
	if src == nil || src.visited || target == nil || target.visited {
		return
	}

	src.visited = true
	target.visited = true
	src.buildTag(srcTag)
	src.buildImportPath()
	target.buildTag(targetTag)
	target.buildImportPath()

	fn := fmt.Sprintf("Convert%s%sTo%s%s", src.importPkg, src.name, target.importPkg, target.name)
	if !outputer.addFuncCheckIfNeedGenerated(fn) {
		return
	}

	fmt.Printf("start renderConvert src: %s target: %s fn: %s\n", src.name, target.name, fn)

	outputer.appendImport(src.importPkg, src.importPath)
	outputer.appendImport(target.importPkg, target.importPath)

	tplCtx := pongo2.Context{
		"SrcFullName":    strings.Title(src.importPkg) + src.name,
		"TargetFullName": strings.Title(target.importPkg) + target.name,
		"SrcPkg":         src.importPkg,
		"SrcName":        src.name,
		"TargetPkg":      target.importPkg,
		"TargetName":     target.name,
	}

	out, err := convertFuncTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}

	outputer.write(out)

	for i := range src.membersByTags {
		tag := src.membersByTags[i].tagName
		sfieldName := src.membersByTags[i].fieldName

		if !token.IsExported(sfieldName) && src.importPkg != "" {
			continue
		}

		smember := src.membersByTags[i].member
		smember.buildImportPath()
		targetMember, exists := target.membersByTagsMap[tag]
		if !exists {
			re := &report{
				fieldName: sfieldName,
				ignore:    true,
				reason:    "Tag Name Mismatch",
			}

			c.appendReport(sfieldName, re)
			continue
		}

		dfieldName := targetMember.fieldName
		dmember := targetMember.member
		dmember.buildImportPath()
		fmt.Printf("render member by tag: %s srcname: %s dstname: %s\n", tag, sfieldName, dfieldName)
		body := c.renderMemberConvert(sfieldName, dfieldName, smember, dmember, outputer)
		if body != "" {
			outputer.write(body)
		}
		re := &report{
			fieldName: sfieldName,
			ignore:    false,
		}

		c.reports = append(c.reports, re)
	}

	outputer.write(copyFuncReturnString)

	// if s.importPath != "" {
	// 	outputer.appendImport(s.importPkg, s.importPath)
	// }

	for i := range c.stack {
		stackC := newConvert()
		stackC.renderConvert(c.stack[i].s, c.stack[i].t, srcTag, targetTag, outputer)
		c.children[c.stack[i].fieldName] = stackC
	}
}

func (c *convert) renderScalarConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer) string {

	var (
		tpl    *pongo2.Template
		tplCtx = pongo2.Context{
			"TargetFieldName": dfieldName,
			"SrcFieldName":    sfieldName,
		}
	)

	switch {
	case (smember.isScalar() && dmember.isScalar()) || (smember.importPkg == "time" && dmember.importPkg == "time"):
		tpl = convertScalarToScalarTemplate

	case (smember.isScalar() && dmember.isStarScalar()) || (smember.importPkg == "time" && dmember.pointerStruct != nil && dmember.pointerStruct.importPkg == "time"):
		tpl = convertScalarToPointerTemplate

	case (smember.isStarScalar() && dmember.isScalar()) || (smember.pointerStruct != nil && smember.pointerStruct.importPkg == "time" && dmember.importPkg == "time"):
		tpl = convertPointerToScalarTemplate

	case (smember.isStarScalar() && dmember.isStarScalar()) || (smember.pointerStruct != nil && smember.pointerStruct.importPkg == "time" && dmember.pointerStruct != nil && dmember.pointerStruct.importPkg == "time"):
		tpl = convertPointerScalarToPointerScalarTemplate

	default:
		re := &report{
			fieldName: sfieldName,
			ignore:    true,
			reason:    "Field Type Mismatch",
		}

		c.appendReport(sfieldName, re)
		return ""
	}

	out, err := tpl.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	return out
}

func (c *convert) renderStructConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer) string {
	var (
		tpl          *pongo2.Template
		tplCtx       pongo2.Context
		targetMember = dmember
		sourceMember = smember
	)

	switch {
	case smember.isStruct() && dmember.isStruct():
		tpl = convertStructToStructTemplate
		tplCtx = pongo2.Context{
			"SrcFullName":     strings.Title(smember.importPkg) + smember.name,
			"TargetFullName":  strings.Title(dmember.importPkg) + dmember.name,
			"TargetPkg":       dmember.importPkg,
			"TargetFieldName": dfieldName,
			"SrcFieldName":    sfieldName,
		}

	case smember.isStruct() && dmember.isStarStruct():
		targetMember = dmember.pointerStruct
		tpl = convertStructToPointerTemplate
		tplCtx = pongo2.Context{
			"SrcFullName":     strings.Title(smember.importPkg) + smember.name,
			"TargetFullName":  strings.Title(targetMember.importPkg) + targetMember.name,
			"TargetPkg":       targetMember.importPkg,
			"TargetFieldName": dfieldName,
			"SrcFieldName":    sfieldName,
		}

	case smember.isStarStruct() && dmember.isStarStruct():
		sourceMember = smember.pointerStruct
		targetMember = dmember.pointerStruct
		tpl = convertPointerStructToPointerStructTemplate
		tplCtx = pongo2.Context{
			"SrcFullName":     strings.Title(sourceMember.importPkg) + sourceMember.name,
			"TargetFullName":  strings.Title(targetMember.importPkg) + targetMember.name,
			"TargetPkg":       targetMember.importPkg,
			"TargetFieldName": dfieldName,
			"SrcFieldName":    sfieldName,
		}

	case smember.isStarStruct() && dmember.isStruct():
		sourceMember = smember.pointerStruct
		tpl = convertPointerToStructTemplate
		tplCtx = pongo2.Context{
			"SrcFullName":     strings.Title(sourceMember.importPkg) + sourceMember.name,
			"TargetFullName":  strings.Title(dmember.importPkg) + dmember.name,
			"TargetFieldName": dfieldName,
			"SrcFieldName":    sfieldName,
		}

	default:
		re := &report{
			fieldName: sfieldName,
			ignore:    true,
			reason:    "Field Type Mismatch",
		}

		c.appendReport(sfieldName, re)
		return ""
	}

	out, err := tpl.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)

	wc := &wrapperCtx{
		s:         sourceMember,
		t:         targetMember,
		fieldName: sfieldName,
	}
	c.stack = append(c.stack, wc)
	return ""
}

func (c *convert) renderArrayScalarConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	tpl := convertSliceScalarTemplate
	tplCtx := pongo2.Context{
		"ArrayLength":     arrayLength,
		"TargetType":      dmember.scalarType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := tpl.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)
	return ""
}

func (c *convert) renderArrayStructConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	targetType := dmember.importPkg + "." + dmember.name
	tplCtx := pongo2.Context{
		"SrcFullName":     strings.Title(smember.importPkg) + smember.name,
		"TargetFullName":  strings.Title(dmember.importPkg) + dmember.name,
		"ArrayLength":     arrayLength,
		"TargetType":      targetType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := convertSlicePointerStructTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)

	wc := &wrapperCtx{
		s:         smember,
		t:         dmember,
		fieldName: sfieldName,
	}
	c.stack = append(c.stack, wc)
	return ""
}

func (c *convert) renderArrayPointerStructConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	targetType := "*" + dmember.importPkg + "." + dmember.name
	tplCtx := pongo2.Context{
		"SrcFullName":     strings.Title(smember.importPkg) + smember.name,
		"TargetFullName":  strings.Title(dmember.importPkg) + dmember.name,
		"ArrayLength":     arrayLength,
		"TargetType":      targetType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := convertSlicePointerStructTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)

	wc := &wrapperCtx{
		s:         smember,
		t:         dmember,
		fieldName: sfieldName,
	}
	c.stack = append(c.stack, wc)
	return ""
}

func (c *convert) renderArrayPointerScalarConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	tpl := convertSlicePointerScalarTemplate
	tplCtx := pongo2.Context{
		"ArrayLength":     arrayLength,
		"TargetType":      dmember.scalarType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := tpl.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)
	return ""
}

func (c *convert) renderArrayScalarToPointerConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	tpl := convertSliceScalarToPointerTemplate
	tplCtx := pongo2.Context{
		"ArrayLength":     arrayLength,
		"TargetType":      dmember.scalarType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := tpl.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)
	return ""
}

func (c *convert) renderArrayPointerToScalarConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	tpl := convertSlicePointerToScalarTemplate
	tplCtx := pongo2.Context{
		"ArrayLength":     arrayLength,
		"TargetType":      dmember.scalarType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := tpl.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)
	return ""
}

func (c *convert) renderArrayPointerToStructConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	targetType := dmember.importPkg + "." + dmember.name
	tplCtx := pongo2.Context{
		"SrcFullName":     strings.Title(smember.importPkg) + smember.name,
		"TargetFullName":  strings.Title(dmember.importPkg) + dmember.name,
		"ArrayLength":     arrayLength,
		"TargetType":      targetType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := convertSlicePointerToStructTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)

	wc := &wrapperCtx{
		s:         smember,
		t:         dmember,
		fieldName: sfieldName,
	}
	c.stack = append(c.stack, wc)
	return ""
}

func (c *convert) renderArrayStructToPointerConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer, arrayLength string) string {
	targetType := "*" + dmember.importPkg + "." + dmember.name
	tplCtx := pongo2.Context{
		"SrcFullName":     strings.Title(smember.importPkg) + smember.name,
		"TargetFullName":  strings.Title(dmember.importPkg) + dmember.name,
		"ArrayLength":     arrayLength,
		"TargetType":      targetType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := convertSliceStructToPointerTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.write(out)

	wc := &wrapperCtx{
		s:         smember,
		t:         dmember,
		fieldName: sfieldName,
	}
	c.stack = append(c.stack, wc)
	return ""
}

func (c *convert) renderArrayConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer) string {
	item := smember.containerItem
	ditem := dmember.containerItem

	if smember.scalarType != "" && dmember.scalarType != "" {
		return c.renderArrayScalarConvert(sfieldName, dfieldName, smember, dmember, outputer, smember.arrayLength)
	}

	// array item is pointer scalar
	if item != nil && item.pointer && item.scalarType != "" && ditem != nil && ditem.pointer && ditem.scalarType != "" {
		return c.renderArrayPointerScalarConvert(sfieldName, dfieldName, item, ditem, outputer, smember.arrayLength)
	}

	// array item src is pointer and target not
	if item != nil && item.pointer && item.scalarType != "" && dmember.scalarType != "" {
		return c.renderArrayPointerToScalarConvert(sfieldName, dfieldName, item, dmember, outputer, smember.arrayLength)
	}

	// array item src is scalar and target is pointer
	if smember.scalarType != "" && ditem != nil && ditem.pointer && ditem.scalarType != "" {
		return c.renderArrayScalarToPointerConvert(sfieldName, dfieldName, smember, ditem, outputer, smember.arrayLength)
	}

	// slice both struct
	if item != nil && item.scalarType == "" && item.pointerStruct == nil && ditem != nil && ditem.scalarType == "" && ditem.pointerStruct == nil {
		return c.renderArrayStructConvert(sfieldName, dfieldName, item, ditem, outputer, smember.arrayLength)
	}

	// slice both pointer struct
	if item != nil && item.scalarType == "" && item.pointerStruct != nil && ditem != nil && ditem.scalarType == "" && ditem.pointerStruct != nil {
		return c.renderArrayPointerStructConvert(sfieldName, dfieldName, item.pointerStruct, ditem.pointerStruct, outputer, smember.arrayLength)
	}

	// slice src is struct and target pointer
	if item != nil && item.scalarType == "" && item.pointerStruct == nil && ditem != nil && ditem.scalarType == "" && ditem.pointerStruct != nil {
		return c.renderArrayStructToPointerConvert(sfieldName, dfieldName, item, ditem.pointerStruct, outputer, smember.arrayLength)
	}

	// slice src is pointer and target is struct
	if item != nil && item.scalarType == "" && item.pointerStruct != nil && ditem != nil && ditem.scalarType == "" && ditem.pointerStruct == nil {
		return c.renderArrayPointerToStructConvert(sfieldName, dfieldName, item.pointerStruct, ditem, outputer, smember.arrayLength)
	}

	re := &report{
		fieldName: sfieldName,
		ignore:    true,
		reason:    "Field Type Mismatch",
	}

	c.appendReport(sfieldName, re)
	return ""
}

func (c *convert) renderMapScalarConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer) string {
	keyType := dmember.containerKey
	valueType := dmember.scalarType

	tplCtx := pongo2.Context{
		"KeyType":         keyType,
		"ValueType":       valueType,
		"TargetFieldName": dfieldName,
		"SrcFieldName":    sfieldName,
	}

	out, err := convertMapScalarTemplate.Execute(tplCtx)
	if err != nil {
		panic(err)
	}
	outputer.writeFunc(out)

	if strings.Contains(dmember.containerKey, ".") {
		outputer.appendImport(dmember.containerPkg, dmember.containerPath)
	}
	return ""
}

func (c *convert) renderMapConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer) string {
	item := smember.containerItem
	ditem := dmember.containerItem

	// map type: map[int]int map[int]string ...
	if smember.scalarType != "" && item == nil && dmember.scalarType != "" && ditem == nil {
		return c.renderMapScalarConvert(sfieldName, dfieldName, smember, dmember, outputer)
	}

	re := &report{
		fieldName: sfieldName,
		ignore:    true,
		reason:    "Field Type Mismatch",
	}

	c.appendReport(sfieldName, re)
	return ""
}

func (c *convert) renderMemberConvert(sfieldName, dfieldName string, smember, dmember *Struct, outputer *outputer) string {
	if sfieldName == "AdditionalServiceTypes" {
		fmt.Printf("AdditionalServiceTypes %+v\n", smember)
	}
	// output scalar type, maybe pointer
	if smember.isScalar() || smember.importPkg == "time" || (smember.pointerStruct != nil && smember.pointerStruct.importPkg == "time") {
		return c.renderScalarConvert(sfieldName, dfieldName, smember, dmember, outputer)
	}

	// normal struct
	if smember.isStruct() {
		return c.renderStructConvert(sfieldName, dfieldName, smember, dmember, outputer)
	}

	// normal pointer struct, not *[]int *map[string]int
	if smember.isStarStruct() {
		return c.renderStructConvert(sfieldName, dfieldName, smember, dmember, outputer)
	}

	// array or slice type
	if smember.isArray() {
		return c.renderArrayConvert(sfieldName, dfieldName, smember, dmember, outputer)
	}

	// map type
	if smember.isMap() {
		return c.renderMapConvert(sfieldName, dfieldName, smember, dmember, outputer)
	}

	re := &report{
		fieldName: sfieldName,
		ignore:    true,
		reason:    "type mismatch",
	}

	c.appendReport(sfieldName, re)
	return ""
}
