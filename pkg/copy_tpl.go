package pkg

import (
	"github.com/flosch/pongo2"
)

var (
	copyFuncTemplate          *pongo2.Template
	copyInterfaceTemplate     *pongo2.Template
	copyScalarTemplate        *pongo2.Template
	copyPointerScalarTemplate *pongo2.Template
	copyStructTemplate        *pongo2.Template
	copyPointerStructTemplate *pongo2.Template

	copySliceScalarTemplate     *pongo2.Template
	copySliceStarScalarTemplate *pongo2.Template
	copySliceStructTemplate     *pongo2.Template
	copySliceStarStructTemplate *pongo2.Template

	copyMapScalarTemplate     *pongo2.Template
	copyMapStarScalarTemplate *pongo2.Template
	copyMapFieldTemplate      *pongo2.Template
	copyMapStructTemplate     *pongo2.Template
	copyMapStarStructTemplate *pongo2.Template
)

func init() {
	copyFuncTemplate = pongo2.Must(pongo2.FromString(copyFuncTemplateString))
	copyInterfaceTemplate = pongo2.Must(pongo2.FromString(copyInterfaceTemplateString))
	copyScalarTemplate = pongo2.Must(pongo2.FromString(copyScalarTemplateString))
	copyPointerScalarTemplate = pongo2.Must(pongo2.FromString(copyPointerScalarTemplateString))
	copyStructTemplate = pongo2.Must(pongo2.FromString(copyStructTemplateString))
	copyPointerStructTemplate = pongo2.Must(pongo2.FromString(copyPointerStructTemplateString))

	copySliceScalarTemplate = pongo2.Must(pongo2.FromString(copySliceScalarTemplateString))
	copySliceStarScalarTemplate = pongo2.Must(pongo2.FromString(copySliceStarScalarTemplateString))
	copySliceStructTemplate = pongo2.Must(pongo2.FromString(copySliceStructTemplateString))
	copySliceStarStructTemplate = pongo2.Must(pongo2.FromString(copySliceStarStructTemplateString))

	copyMapScalarTemplate = pongo2.Must(pongo2.FromString(copyMapScalarTemplateString))
	copyMapStarScalarTemplate = pongo2.Must(pongo2.FromString(copyMapStarScalarTemplateString))
	copyMapFieldTemplate = pongo2.Must(pongo2.FromString(copyMapFieldTemplateString))
	copyMapStructTemplate = pongo2.Must(pongo2.FromString(copyMapStructTemplateString))
	copyMapStarStructTemplate = pongo2.Must(pongo2.FromString(copyMapStarStructTemplateString))
}

var copyFuncReturnString = `    return dst
}
`

var copyInterfaceTemplateString = `    dst.{{ FieldName }} = deepcopy.Copy(src.{{ FieldName }}).({{ FullItemType }})
`

var copyScalarTemplateString = `    dst.{{ FieldName }} = src.{{ FieldName }}
`
var copyPointerScalarTemplateString = `    if src.{{ FieldName }} != nil {
        tmp := *(src.{{ FieldName }})
        dst.{{ FieldName }} = &tmp 
    }
`

var copyFuncTemplateString = `// Copy{{ PkgName}}{{ ItemType }} only copy exported fields
func Copy{{ PkgName}}{{ ItemType }}(src *{{ FullItemType }}) *{{ FullItemType }} {
    if src == nil {
        return nil
    }
    dst := &{{ FullItemType }}{}
`

var copyStructTemplateString = `    tmp{{ FieldName }} := Copy{{ PkgName}}{{ ItemType }}(&src.{{ FieldName }})
    if tmp{{ FieldName }}  != nil {
        dst.{{ FieldName }}  = *tmp{{ FieldName }}
    }
`

var copyPointerStructTemplateString = `    dst.{{ FieldName }} = Copy{{ PkgName}}{{ ItemType }}(src.{{ FieldName }})
`

var copySliceScalarTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ FieldName }} = make([]{{ PkgName}}{{ ItemType }}, len(src.{{ FieldName }}))
    {% endif %}
    for i := range src.{{ FieldName }} {
    	dst.{{ FieldName }}[i] = src.{{ FieldName }}[i]
    }
`

var copySliceStarScalarTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ FieldName }} = make([]{{ PkgName}}{{ ItemType }}, len(src.{{ FieldName }}))
    {% endif %}
    for i := range src.{{ FieldName }} {
    	if src.{{ FieldName }}[i] == nil {
    		continue
    	}
    	tmp := *src.{{ FieldName }}[i]
    	dst.{{ FieldName }}[i] = &tmp
    }
`

var copySliceStructTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ FieldName }} = make([]{{ PkgName}}.{{ ItemType }}, len(src.{{ FieldName }}))
    {% endif %}
    for i := range src.{{ FieldName }} {
    	cp := {{ StructCopyFunc }}(&((src.{{ FieldName }})[i]))
    	if cp == nil {
    		continue
    	}

    	dst.{{ FieldName }}[i] = *cp
    }
`

var copySliceStarStructTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ FieldName }} = make([]*{{ PkgName}}.{{ ItemType }}, len(src.{{ FieldName }}))
    {% endif %}
    for i := range src.{{ FieldName }} {
    	dst.{{ FieldName }}[i] = {{ StructCopyFunc }}(&((src.{{ FieldName }})[i]))
    }
`

var copyMapFieldTemplateString = `
	dst.{{FieldName}} = {{ FuncName }}(src.{{FieldName}})
`

var copyMapScalarTemplateString = `// {{ FuncName }} copy map {{ KeyType }} {{ ValueType }}
func {{ FuncName }}(src map[{{ KeyType }}]{{ ValueType }}) map[{{ KeyType }}]{{ ValueType }} {
	dst:=make(map[{{ KeyType }}]{{ ValueType }}, len(src))
	for key := range src {
		dst[key] = src[key]
	}
	return dst
}
`

var copyMapStarScalarTemplateString = `// {{ FuncName }} copy map {{ KeyType }} {{ ValueType }}
func {{ FuncName }}(src map[{{ KeyType }}]*{{ ValueType }}) map[{{ KeyType }}]*{{ ValueType }} {
	dst:=make(map[{{ KeyType }}]*{{ ValueType }}, len(src))
	for key := range src {
		if src[key] == nil {
			dst[key] = nil
			continue
		}
		tmp := *(src[key])
		dst[key] = &tmp
	}
	return dst
}
`

var copyMapStructTemplateString = `
func {{ FuncName }}(src map[{{ KeyType }}]{{ ValueType }}) map[{{ KeyType }}]{{ ValueType }} {
	if src == nil {
		return nil
	}
	dst:=make(map[{{ KeyType }}]{{ ValueType }}, len(src))
	for key := range src {
		srcValue := src[key]
		dst[key] := *{{ CopyStructFunc }}(&srcValue)
	}
	return dst
}
`

var copyMapStarStructTemplateString = `
func {{ FuncName }}(src map[{{ KeyType }}]*{{ ValueType }}) map[{{ KeyType }}]*{{ ValueType }} {
	if src == nil {
		return nil
	}
	dst:=make(map[{{ KeyType }}]*{{ ValueType }}, len(src))
	for key := range src {
		dst[key] := {{ CopyStructFunc }}(src[key])
	}
	return dst
}
`
