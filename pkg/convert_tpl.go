package pkg

import (
	"github.com/flosch/pongo2"
)

var (
	convertFuncTemplate                         *pongo2.Template
	convertScalarToScalarTemplate               *pongo2.Template
	convertScalarToPointerTemplate              *pongo2.Template
	convertPointerToScalarTemplate              *pongo2.Template
	convertPointerScalarToPointerScalarTemplate *pongo2.Template

	convertStructToStructTemplate               *pongo2.Template
	convertStructToPointerTemplate              *pongo2.Template
	convertPointerToStructTemplate              *pongo2.Template
	convertPointerStructToPointerStructTemplate *pongo2.Template

	convertSliceScalarTemplate          *pongo2.Template
	convertSlicePointerToScalarTemplate *pongo2.Template
	convertSliceScalarToPointerTemplate *pongo2.Template
	convertSlicePointerScalarTemplate   *pongo2.Template

	convertSliceStructTemplate          *pongo2.Template
	convertSlicePointerStructTemplate   *pongo2.Template
	convertSliceStructToPointerTemplate *pongo2.Template
	convertSlicePointerToStructTemplate *pongo2.Template

	convertMapScalarTemplate *pongo2.Template
)

func init() {
	convertFuncTemplate = pongo2.Must(pongo2.FromString(convertFuncTemplateString))

	convertScalarToScalarTemplate = pongo2.Must(pongo2.FromString(convertScalarToScalarTemplateString))
	convertScalarToPointerTemplate = pongo2.Must(pongo2.FromString(convertScalarToPointerTemplateString))
	convertPointerToScalarTemplate = pongo2.Must(pongo2.FromString(convertPointerToScalarTemplateString))
	convertPointerScalarToPointerScalarTemplate = pongo2.Must(pongo2.FromString(convertPointerScalarToPointerScalarTemplateString))

	convertStructToStructTemplate = pongo2.Must(pongo2.FromString(convertStructToStructTemplateString))
	convertStructToPointerTemplate = pongo2.Must(pongo2.FromString(convertStructToPointerTemplateString))
	convertPointerToStructTemplate = pongo2.Must(pongo2.FromString(convertPointerToStructTemplateString))
	convertPointerStructToPointerStructTemplate = pongo2.Must(pongo2.FromString(convertPointerStructToPointerStructTemplateString))

	convertSliceScalarTemplate = pongo2.Must(pongo2.FromString(convertSliceScalarTemplateString))
	convertSlicePointerScalarTemplate = pongo2.Must(pongo2.FromString(convertSlicePointerScalarTemplateString))
	convertSlicePointerToScalarTemplate = pongo2.Must(pongo2.FromString(convertSlicePointerToScalarTemplateString))
	convertSliceScalarToPointerTemplate = pongo2.Must(pongo2.FromString(convertSliceScalarToPointerTemplateString))
	convertSliceStructTemplate = pongo2.Must(pongo2.FromString(convertSliceStructTemplateString))
	convertSlicePointerStructTemplate = pongo2.Must(pongo2.FromString(convertSlicePointerStructTemplateString))
	convertSliceStructToPointerTemplate = pongo2.Must(pongo2.FromString(convertSliceStructToPointerTemplateString))
	convertSlicePointerToStructTemplate = pongo2.Must(pongo2.FromString(convertSlicePointerToStructTemplateString))

	convertMapScalarTemplate = pongo2.Must(pongo2.FromString(convertMapScalarTemplateString))
}

var convertFuncTemplateString = `// Convert{{ SrcFullName}}To{{ TargetFullName }} only convert exported fields
func Convert{{ SrcFullName}}To{{ TargetFullName }}(src *{{ SrcPkg }}.{{ SrcName }}) *{{ TargetPkg }}.{{ TargetName }} {
    if src == nil {
        return nil
    }
    dst := &{{ TargetPkg }}.{{ TargetName }}{}
`

var convertFuncReturnString = `    return dst
}
`

// ------------------scalar related
var convertScalarToScalarTemplateString = `    dst.{{ TargetFieldName }} = src.{{ SrcFieldName }}
`

var convertPointerToScalarTemplateString = `    if src.{{ SrcFieldName }} != nil {
        dst.{{ TargetFieldName }} = *(src.{{ SrcFieldName }})
    }
`

var convertScalarToPointerTemplateString = `    tmp{{ TargetFieldName }} := src.{{ SrcFieldName }}
    dst.{{ TargetFieldName }} = &tmp{{ TargetFieldName }} 
`

var convertPointerScalarToPointerScalarTemplateString = `    if src.{{ SrcFieldName }} != nil {
        tmp := *(src.{{ SrcFieldName }})
        dst.{{ TargetFieldName }} = &tmp 
    }
`

// ------------------scalar end

// ------------------struct related
var convertStructToStructTemplateString = `    star{{ TargetFieldName }} := Convert{{ SrcFullName}}To{{ TargetFullName }}(&src.{{ SrcFieldName }})
    if star{{ TargetFieldName }}  != nil {
        dst.{{ TargetFieldName }}  = *star{{ TargetFieldName }}
    }
`

var convertStructToPointerTemplateString = `    dst.{{ TargetFieldName }} = Convert{{ SrcFullName}}To{{ TargetFullName }}(&src.{{ SrcFieldName }})
`

var convertPointerToStructTemplateString = `    star{{ TargetFieldName }} := Convert{{ SrcFullName}}To{{ TargetFullName }}(src.{{ SrcFieldName }})
    if star{{ TargetFieldName }}  != nil {
        dst.{{ TargetFieldName }}  = *star{{ TargetFieldName }}
    }
`

var convertPointerStructToPointerStructTemplateString = `    dst.{{ TargetFieldName }} = Convert{{ SrcFullName}}To{{ TargetFullName }}(src.{{ SrcFieldName }})
`

// ------------------struct end

// ------------------array related
var convertSliceScalarTemplateString = `    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	dst.{{ TargetFieldName }}[i] = src.{{ SrcFieldName }}[i]
    }
`

var convertSlicePointerScalarTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	if src.{{ SrcFieldName }}[i] == nil {
    		continue
    	}

    	tmp := *src.{{ SrcFieldName }}[i] 
    	dst.{{ TargetFieldName }}[i] = &tmp
    }
`

var convertSlicePointerToScalarTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	if src.{{ SrcFieldName }}[i] == nil {
    		continue
    	}

    	dst.{{ TargetFieldName }}[i] = *src.{{ SrcFieldName }}[i] 
    }
`

var convertSliceScalarToPointerTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	if src.{{ SrcFieldName }}[i] == nil {
    		continue
    	}

    	tmp := src.{{ SrcFieldName }}[i] 
    	dst.{{ TargetFieldName }}[i] = &tmp
    }
`

var convertSlicePointerStructTemplateString = `    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	dst.{{ TargetFieldName }}[i] = Convert{{ SrcFullName}}To{{ TargetFullName }}(src.{{ SrcFieldName }}[i])
    }
`

var convertSliceStructTemplateString = `
    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	tmp := Convert{{ SrcFullName}}To{{ TargetFullName }}(&src.{{ SrcFieldName }}[i])
    	if tmp != nil {
            dst.{{ TargetFieldName }}[i] = &tmp
    	}
    }
`

var convertSlicePointerToStructTemplateString = `    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	tmp := Convert{{ SrcFullName}}To{{ TargetFullName }}(src.{{ SrcFieldName }}[i])
    	if tmp != nil {
    		dst.{{ TargetFieldName }}[i] = *tmp
    	}
    }
`

var convertSliceStructToPointerTemplateString = `    {% if ArrayLength == "" %}
    dst.{{ TargetFieldName }} = make([]{{ TargetType }}, len(src.{{ SrcFieldName }}))
    {% endif %}
    for i := range src.{{ SrcFieldName }} {
    	dst.{{ TargetFieldName }}[i] = Convert{{ SrcFullName}}To{{ TargetFullName }}(&src.{{ SrcFieldName }}[i])
    }
`

// ------------------array end

// ------------------map related
var convertMapScalarTemplateString = `    dst.{{ TargetFieldName }} = make(map[{{ KeyType }}]{{ ValueType }}, len(src.{{ SrcFieldName }}))
    for i := range src.{{ SrcFieldName }} {
    	dst.{{ TargetFieldName }}[i] = src.{{ SrcFieldName }}[i]
    }
`

// ------------------map end
