package pkg

import (
	"github.com/flosch/pongo2"
)

var (
	funcTemplate              *pongo2.Template
	interfaceTemplate         *pongo2.Template
	pointerTemplate           *pongo2.Template
	scalarTemplate            *pongo2.Template
	diffStructTemplate        *pongo2.Template
	diffPointerStructTemplate *pongo2.Template

	floatCompareTemplate     *pongo2.Template
	floatCompareFuncTemplate *pongo2.Template
	pointerFloatTemplate     *pongo2.Template

	// pointer related
	pointerDiffTemplate        *pongo2.Template
	pointerSliceScalarTemplate *pongo2.Template
	pointerArrayScalarTemplate *pongo2.Template
	diffPointerArrayTemplate   *pongo2.Template

	// slice related
	slicePointerStructTemplate  *pongo2.Template
	slicePointerScalarTemplate  *pongo2.Template
	sliceScalarTemplate         *pongo2.Template
	sliceStructTemplate         *pongo2.Template
	diffStarSliceScalarTemplate *pongo2.Template

	// map related
	mapPointerScalarTemplate *pongo2.Template
	mapScalarTemplate        *pongo2.Template
	mapStructTemplate        *pongo2.Template
	mapPointerStructTemplate *pongo2.Template

	// time compare
	isTimeInDeltaTemplate *pongo2.Template
	diffTimeTemplate      *pongo2.Template
)

func init() {
	funcTemplate = pongo2.Must(pongo2.FromString(funcTemplateString))
	scalarTemplate = pongo2.Must(pongo2.FromString(scalarTemplateString))
	interfaceTemplate = pongo2.Must(pongo2.FromString(interfaceTemplateString))
	diffStructTemplate = pongo2.Must(pongo2.FromString(diffStructTemplateString))
	diffPointerStructTemplate = pongo2.Must(pongo2.FromString(diffPointerStructTemplateString))
	pointerTemplate = pongo2.Must(pongo2.FromString(pointerTemplateString))
	pointerDiffTemplate = pongo2.Must(pongo2.FromString(pointerDiffTemplateString))

	floatCompareTemplate = pongo2.Must(pongo2.FromString(floatCompareTemplateString))
	floatCompareFuncTemplate = pongo2.Must(pongo2.FromString(floatCompareFuncTemplateString))
	pointerFloatTemplate = pongo2.Must(pongo2.FromString(pointerFloatTemplateString))
	pointerSliceScalarTemplate = pongo2.Must(pongo2.FromString(pointerSliceScalarTemplateString))
	pointerArrayScalarTemplate = pongo2.Must(pongo2.FromString(pointerArrayScalarTemplateString))
	diffPointerArrayTemplate = pongo2.Must(pongo2.FromString(diffPointerArrayTemplateString))

	// slice related
	sliceStructTemplate = pongo2.Must(pongo2.FromString(sliceStructTemplateString))
	slicePointerStructTemplate = pongo2.Must(pongo2.FromString(slicePointerStructTemplateString))
	slicePointerScalarTemplate = pongo2.Must(pongo2.FromString(slicePointerScalarTemplateString))
	sliceScalarTemplate = pongo2.Must(pongo2.FromString(sliceScalarTemplateString))
	diffStarSliceScalarTemplate = pongo2.Must(pongo2.FromString(diffStarSliceScalarTemplateString))

	// map related
	mapPointerScalarTemplate = pongo2.Must(pongo2.FromString(mapPointerScalarTemplateString))
	mapScalarTemplate = pongo2.Must(pongo2.FromString(mapScalarTemplateString))
	mapStructTemplate = pongo2.Must(pongo2.FromString(mapStructTemplateString))
	mapPointerStructTemplate = pongo2.Must(pongo2.FromString(mapPointerStructTemplateString))

	// time
	isTimeInDeltaTemplate = pongo2.Must(pongo2.FromString(isTimeInDeltaTemplateString))
	diffTimeTemplate = pongo2.Must(pongo2.FromString(diffTimeTemplateString))
}

var diffTimeTemplateString = `    if !IsTimeInDelta(x.{{ FieldName }}, y.{{ FieldName }}) {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
`

var scalarTemplateString = `    if x.{{ FieldName }} != y.{{ FieldName }} {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
`

var floatCompareFuncTemplateString = `    if ! Is{{ ItemType }}InDelta(x.{{ FieldName }}, y.{{ FieldName }}) {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
`

// 结构体名，如果有 import alias 就用 alias, 没有就用 import pkg 名
// Diff_[First.Name]_[Second.Name]_[Last.Name] 这样子用于区分
// Diff[prefix][fnname]
var diffTemplate = `    if diff := Diff%s%s(&%s, &%s); diff != "" {
        return "%s." + diff
    }
`

var diffStructTemplateString = `    if diff := Diff{{ Prefix }}{{ FullItemType }}(&(x.{{ FieldName}}), &(y.{{ FieldName}})); diff != "" {
        return "%s." + diff
    }
`

var diffPointerStructTemplateString = `    if diff := Diff{{ Prefix }}{{ FullItemType }}(x.{{ FieldName}}, y.{{ FieldName}}); diff != "" {
        return "%s." + diff
    }
`

var diffPointerTemplate = `    if diff := Diff%s%s%s(%s, %s); diff != "" {
        return "%s." + diff
    }
`

var diffPointerArrayTemplateString = `    if diff:={{ DiffFunc }}(x.{{ FieldName }}, y.{{ FieldName }}); diff != "" {
        return "%s." + diff
    }
`

var diffPointerUtilTemplate = `    if diff := %s(x.%s, y.%s); diff != "" {
        return "%s." + diff
    }
`

var diffUtilTemplate = `    if diff := %s(&x.%s, &y.%s); diff != "" {
        return "%s." + diff
    }
`

var returnEmpty = `
    return ""
}

`

var firstPointer = `    if x == y {
        return ""
    }
    if x == nil || y == nil {
        return fmt.Sprintf("%s got:%%v expected:%%v", x, y)
    }
`

var pointerBase1 = `    if (x.%s == nil && y.%s != nil ) || (x.%s != nil && y.%s == nil) {
        return fmt.Sprintf("%s got:%%v expected:%%v", x.%s, y.%s)
    }
`
var pointerBase2 = `    if *x.%s != *y.%s {
        return fmt.Sprintf("%s got:%%v expected:%%v", x.%s, y.%s)
    }
`

var pointerDiffTemplateString = `    if x.{{ FieldName }} != y.{{ FieldName }} {
        return fmt.Sprintf("%s got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
    if x.{{ FieldName }} == nil || y.{{ FieldName }} == nil {
        return fmt.Sprintf("%s got:%v expected:%v", x, y)
    }
`

var pointerTemplateString = `    if (x.{{ FieldName }} == nil && y.{{ FieldName }} != nil ) || (x.{{ FieldName }} != nil && y.{{ FieldName }} == nil) {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
    if (x.{{ FieldName }} != nil && y.{{ FieldName }} != nil ) && (*x.{{ FieldName }} != *y.{{ FieldName }}) {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
`

var pointerFloatTemplateString = `    if (x.{{ FieldName }} == nil && y.{{ FieldName }} != nil ) || (x.{{ FieldName }} != nil && y.{{ FieldName }} == nil) {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
    if (x.{{ FieldName }} != nil && y.{{ FieldName }} != nil ) && ! Is{{ ItemType }}InDelta(*x.{{ FieldName }}, *y.{{ FieldName }}) {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
`

var funcSignature = `// %s diff field %s
func %s(x, y *%s) string {
`

var funcTemplateString = `// {{ FuncName }} diff field {{ FieldName }}
func {{ FuncName }}(x, y *{{ FieldType }}) string {
    if x == y {
        return ""
    }
    if x == nil || y == nil {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x, y)
    }
`

var interfaceTemplateString = `    if !reflect.DeepEqual(x.{{ FieldName }}, y.{{ FieldName }}) {
        return fmt.Sprintf("{{ FieldName }} got:%v expected:%v", x.{{ FieldName }}, y.{{ FieldName }})
    }
`

var containerLength = ` if len(%s) != len(%s) {
    return fmt.Sprintf("%s got:%%v expected:%%v", x.%s, y.%s)
}
`

var containerPointer = ` if unsafe.Pointer(&%s) != unsafe.Pointer(&%s) {
    return fmt.Sprintf("%s got:%%v expected:%%v", x.%s, y.%s)
}
`

// below are slice related
var diffStarSliceScalarTemplateString = `// func diffStarSlice{{ PkgName}}{{ ItemType }}
func diffStarSlice{{ Star }}{{ PkgName}}{{ ItemType }}(x, y *[]{{ FullItemType }}) string {
    if x == y {
        return ""
    }
    if x == nil || y == nil {
        return fmt.Sprintf("%s got:%%v expected:%%v", x, y)
    }
    return diffSlice{{Star}}{{ FullItemType }}(*x, *y)
}
`

var sliceScalarTemplateString = `// diffSlice{{ ArrayLength }}{{ ItemType }} ...
func diffSlice{{ ArrayLength }}{{ PkgName}}{{ ItemType }}(x, y [{{ ArrayLength }}]{{ FullItemType }}) string {
    {% if ArrayLength == "" %}
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    {% endif %}
    for i := range x {
        if x[i] != y[i] {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var pointerArrayScalarTemplateString = `// diffStarArray{{ ArrayLength }}{{ Star }}{{ ItemType }} ...
func diffStarArray{{ ArrayLength }}{{ Star }}{{ ItemType }}(x, y *[{{ ArrayLength }}]{{ FullItemType }}) string {
    if x == y {
        return ""
    }
    if x == nil || y == nil {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    {% if Star == "Star" %}
    for i := range *x {
        xvalue := (*x)[i]
        yvalue := (*y)[i]
        if xvalue == yvalue {
            continue
        }
        if xvalue == nil || yvalue == nil {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
        if *xvalue != *yvalue {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    {% else %}
    for i := range *x {
        if (*x)[i] != (*y)[i] {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    {% endif %}
    return ""
}
`

var pointerSliceScalarTemplateString = `// diffStarSlice{{ Star }}{{ ItemType }} ...
func diffStarSlice{{ Star }}{{ ItemType }}(x, y *[]{{ FullItemType }}) string {
    if x == y {
        return ""
    }
    if x == nil || y == nil {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    if diff := diffSlice{{ Star }}{{ ItemType }}(*x, *y); diff != "" {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    return ""
}
`

var slicePointerScalarTemplateString = `// diffSlice{{ ArrayLength }}Star{{ PkgName}}{{ ItemType }} ...
func diffSlice{{ ArrayLength }}Star{{ PkgName}}{{ ItemType }}(x, y [{{ ArrayLength }}]{{ Star }}{{ FullItemType }}) string {
    {% if ArrayLength == "" %}
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    {% endif %}
    for i := range x {
        if x[i] == y[i] {
            continue
        }
        if (x[i] == nil || y[i] == nil) || *(x[i]) != *(y[i]) {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var sliceStructTemplateString = `// diffSlice{{ ArrayLength }}{{ Star }}{{ ItemType }} ...
func diffSlice{{ ArrayLength }}{{ PkgName}}{{ ItemType }}(x, y [{{ ArrayLength }}]{{ Star }}{{ FullItemType }}) string {
    {% if ArrayLength == "" %}
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    {% endif %}
    for i := range x {
        if diff := {{ StructDiffFunc }}(&x[i], &y[i]); diff != "" {
            return fmt.Sprintf("{{ ItemType }} got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var slicePointerStructTemplateString = `// diffSlice{{ ArrayLength }}{{ Star }}{{ ItemType }} ...
func diffSlice{{ ArrayLength }}Star{{ PkgName}}{{ ItemType }}(x, y [{{ ArrayLength }}]{{ Star }}{{ FullItemType }}) string {
    {% if ArrayLength == "" %}
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    {% endif %}
    for i := range x {
        if diff := {{ StructDiffFunc }}(x[i], y[i]); diff != "" {
            return fmt.Sprintf("{{ ItemType }} got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var sliceInterfaceTemplateString = `// diffSlice{{ ArrayLength }}{{ Star }}{{ ItemType }} ...
func diffSlice{{ ArrayLength }}{{ PkgName}}{{ ItemType }}(x, y [{{ ArrayLength }}]{{ Star }}{{ FullItemType }}) string {
    {% if ArrayLength == "" %}
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    {% endif %}
    for i := range x {
        if reflect.DeepEqual(x[i], y[i]) {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

// below are map related

var mapScalarTemplateString = `// {{ FuncName }} ...
func {{ FuncName }}(x, y map[{{ KeyType }}]{{ ValueType }}) string {
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    for key := range x {
        value, exists := y[key]
        if !exists || x[key] != value {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var mapPointerScalarTemplateString = `// {{ FuncName }} ...
func {{ FuncName }}(x, y map[{{ KeyType }}]{{ Star }}{{ ValueType }}) string {
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    for key := range x {
        _, exists := y[key]
        if !exists {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
        if x[key] == y[key] {
            continue
        }
        if (x[key] == nil || y[key] == nil) || *(x[key]) != *(y[key]) {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var mapStructTemplateString = `// {{ FuncName }} ...
func {{ FuncName }}(x, y map[{{ KeyType }}]{{ ValueType }}) string {
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    for key := range x {
        value, exists := y[key]
        if !exists {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }

        xValue := x[key]
        if {{ StructDiffFunc }}(&xValue, &value) != "" {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var mapPointerStructTemplateString = `// {{ FuncName }} ...
func {{ FuncName }}(x, y map[{{ KeyType }}]*{{ ValueType }}) string {
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    for key := range x {
        value, exists := y[key]
        if !exists || {{ StructDiffFunc }}(x[key], value) != "" {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var mapInterfaceTemplateString = `// {{ FuncName }} ...
func {{ FuncName }}(x, y map[{{ KeyType }}]{{ ValueType }}) string {
    if x == nil && y == nil {
        return ""
    }
    if unsafe.Pointer(&x) == unsafe.Pointer(&y) {
        return ""
    }
    if len(x) != len(y) {
        return fmt.Sprintf(" got:%v expected:%v", x, y)
    }
    for key := range x {
        value, exists := y[key]
        if !exists || !reflect.DeepEqual(x[key], value) {
            return fmt.Sprintf(" got:%v expected:%v", x, y)
        }
    }
    return ""
}
`

var floatCompareTemplateString = `// Is{{ ItemType }}InDelta compare float delta in 1e-6
func Is{{ ItemType }}InDelta(x, y {{ ItemType }}) bool {
    if math.IsNaN(x) && math.IsNaN(y) {
        return true
    }

    if (math.IsNaN(x) && !math.IsNaN(y)) || (!math.IsNaN(x) && math.IsNaN(y)) {
        return false
    }

    delta := 1e-6
    dt := x - y
    if dt < -delta || dt > delta {
        return false
    }

    return true
}
`

var isTimeInDeltaTemplateString = `// IsTimeInDelta ...
func IsTimeInDelta(x, y time.Time) bool {
    if x.IsZero() && y.IsZero() {
        return true
    }

    if x.IsZero() || y.IsZero() {
        return false
    }

    if x.After(y) {
        x, y = y, x
    }

    // todo:
    // make it configurable
    margin := 1 * time.Minute
    return !x.Add(margin).Before(y)
}
`
