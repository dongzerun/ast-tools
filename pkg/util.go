package pkg

import (
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

var scalarTypes = map[string]struct{}{
	"int8":       {},
	"uint8":      {},
	"byte":       {},
	"bool":       {},
	"int16":      {},
	"uint16":     {},
	"int32":      {},
	"rune":       {},
	"uint32":     {},
	"int64":      {},
	"uint64":     {},
	"int":        {},
	"uint":       {},
	"uintptr":    {},
	"float32":    {},
	"float64":    {},
	"complex64":  {},
	"complex128": {},
	"string":     {},
	"error":      {}, // error as scalar type
}

var floatTypes = map[string]struct{}{
	"float32": {},
	"float64": {},
}

var goLibraryTypes = map[string]struct{}{
	"time.Time":       {},
	"time.Duration":   {},
	"http.Request":    {},
	"sync.RWMutex":    {},
	"sync.Mutext":     {},
	"sync.Cond":       {},
	"sync.Once":       {},
	"context.Context": {},
}

// goLibraries contain all source libaray import paths
var goLibraries = map[string]struct{}{
	"\"archive\"":              {},
	"\"archive/tar\"":          {},
	"\"archive/zip\"":          {},
	"\"bufio\"":                {},
	"\"builtin\"":              {},
	"\"bytes\"":                {},
	"\"cmd\"":                  {},
	"\"cmd/addr2line\"":        {},
	"\"cmd/api\"":              {},
	"\"cmd/asm\"":              {},
	"\"cmd/buildid\"":          {},
	"\"cmd/cgo\"":              {},
	"\"cmd/compile\"":          {},
	"\"cmd/cover\"":            {},
	"\"cmd/dist\"":             {},
	"\"cmd/doc\"":              {},
	"\"cmd/fix\"":              {},
	"\"cmd/go\"":               {},
	"\"cmd/gofmt\"":            {},
	"\"cmd/internal\"":         {},
	"\"cmd/link\"":             {},
	"\"cmd/nm\"":               {},
	"\"cmd/objdump\"":          {},
	"\"cmd/pack\"":             {},
	"\"cmd/pprof\"":            {},
	"\"cmd/test2json\"":        {},
	"\"cmd/trace\"":            {},
	"\"cmd/vendor\"":           {},
	"\"cmd/vet\"":              {},
	"\"compress\"":             {},
	"\"compress/bzip2\"":       {},
	"\"compress/flate\"":       {},
	"\"compress/gzip\"":        {},
	"\"compress/lzw\"":         {},
	"\"compress/testdata\"":    {},
	"\"compress/zlib\"":        {},
	"\"constraints\"":          {},
	"\"container\"":            {},
	"\"container/heap\"":       {},
	"\"container/list\"":       {},
	"\"container/ring\"":       {},
	"\"context\"":              {},
	"\"crypto\"":               {},
	"\"crypto/aes\"":           {},
	"\"crypto/cipher\"":        {},
	"\"crypto/des\"":           {},
	"\"crypto/dsa\"":           {},
	"\"crypto/ecdsa\"":         {},
	"\"crypto/ed25519\"":       {},
	"\"crypto/elliptic\"":      {},
	"\"crypto/hmac\"":          {},
	"\"crypto/internal\"":      {},
	"\"crypto/md5\"":           {},
	"\"crypto/rand\"":          {},
	"\"crypto/rc4\"":           {},
	"\"crypto/rsa\"":           {},
	"\"crypto/sha1\"":          {},
	"\"crypto/sha256\"":        {},
	"\"crypto/sha512\"":        {},
	"\"crypto/subtle\"":        {},
	"\"crypto/tls\"":           {},
	"\"crypto/x509\"":          {},
	"\"database\"":             {},
	"\"database/sql\"":         {},
	"\"debug\"":                {},
	"\"debug/buildinfo\"":      {},
	"\"debug/dwarf\"":          {},
	"\"debug/elf\"":            {},
	"\"debug/gosym\"":          {},
	"\"debug/macho\"":          {},
	"\"debug/pe\"":             {},
	"\"debug/plan9obj\"":       {},
	"\"embed\"":                {},
	"\"embed/internal\"":       {},
	"\"encoding\"":             {},
	"\"encoding/ascii85\"":     {},
	"\"encoding/asn1\"":        {},
	"\"encoding/base32\"":      {},
	"\"encoding/base64\"":      {},
	"\"encoding/binary\"":      {},
	"\"encoding/csv\"":         {},
	"\"encoding/gob\"":         {},
	"\"encoding/hex\"":         {},
	"\"encoding/json\"":        {},
	"\"encoding/pem\"":         {},
	"\"encoding/xml\"":         {},
	"\"errors\"":               {},
	"\"expvar\"":               {},
	"\"flag\"":                 {},
	"\"fmt\"":                  {},
	"\"go\"":                   {},
	"\"go/ast\"":               {},
	"\"go/build\"":             {},
	"\"go/constant\"":          {},
	"\"go/doc\"":               {},
	"\"go/format\"":            {},
	"\"go/importer\"":          {},
	"\"go/internal\"":          {},
	"\"go/parser\"":            {},
	"\"go/printer\"":           {},
	"\"go/scanner\"":           {},
	"\"go/token\"":             {},
	"\"go/types\"":             {},
	"\"hash\"":                 {},
	"\"hash/adler32\"":         {},
	"\"hash/crc32\"":           {},
	"\"hash/crc64\"":           {},
	"\"hash/fnv\"":             {},
	"\"hash/maphash\"":         {},
	"\"html\"":                 {},
	"\"html/template\"":        {},
	"\"image\"":                {},
	"\"image/color\"":          {},
	"\"image/draw\"":           {},
	"\"image/gif\"":            {},
	"\"image/internal\"":       {},
	"\"image/jpeg\"":           {},
	"\"image/png\"":            {},
	"\"image/testdata\"":       {},
	"\"index\"":                {},
	"\"index/suffixarray\"":    {},
	"\"io\"":                   {},
	"\"io/fs\"":                {},
	"\"io/ioutil\"":            {},
	"\"log\"":                  {},
	"\"log/syslog\"":           {},
	"\"math\"":                 {},
	"\"math/big\"":             {},
	"\"math/bits\"":            {},
	"\"math/cmplx\"":           {},
	"\"math/rand\"":            {},
	"\"mime\"":                 {},
	"\"mime/multipart\"":       {},
	"\"mime/quotedprintable\"": {},
	"\"mime/testdata\"":        {},
	"\"net\"":                  {},
	"\"net/http\"":             {},
	"\"net/internal\"":         {},
	"\"net/mail\"":             {},
	"\"net/netip\"":            {},
	"\"net/rpc\"":              {},
	"\"net/smtp\"":             {},
	"\"net/testdata\"":         {},
	"\"net/textproto\"":        {},
	"\"net/url\"":              {},
	"\"os\"":                   {},
	"\"os/exec\"":              {},
	"\"os/signal\"":            {},
	"\"os/testdata\"":          {},
	"\"os/user\"":              {},
	"\"path\"":                 {},
	"\"path/filepath\"":        {},
	"\"plugin\"":               {},
	"\"reflect\"":              {},
	"\"reflect/internal\"":     {},
	"\"regexp\"":               {},
	"\"regexp/syntax\"":        {},
	"\"regexp/testdata\"":      {},
	"\"runtime\"":              {},
	"\"runtime/asan\"":         {},
	"\"runtime/cgo\"":          {},
	"\"runtime/debug\"":        {},
	"\"runtime/internal\"":     {},
	"\"runtime/metrics\"":      {},
	"\"runtime/msan\"":         {},
	"\"runtime/pprof\"":        {},
	"\"runtime/race\"":         {},
	"\"runtime/testdata\"":     {},
	"\"runtime/trace\"":        {},
	"\"sort\"":                 {},
	"\"strconv\"":              {},
	"\"strconv/testdata\"":     {},
	"\"strings\"":              {},
	"\"sync\"":                 {},
	"\"sync/atomic\"":          {},
	"\"syscall\"":              {},
	"\"syscall/js\"":           {},
	"\"testdata\"":             {},
	"\"testing\"":              {},
	"\"testing/fstest\"":       {},
	"\"testing/internal\"":     {},
	"\"testing/iotest\"":       {},
	"\"testing/quick\"":        {},
	"\"text\"":                 {},
	"\"text/scanner\"":         {},
	"\"text/tabwriter\"":       {},
	"\"text/template\"":        {},
	"\"time\"":                 {},
	"\"time/testdata\"":        {},
	"\"time/tzdata\"":          {},
	"\"unicode\"":              {},
	"\"unicode/utf16\"":        {},
	"\"unicode/utf8\"":         {},
	"\"unsafe\"":               {},
}

func getTagFromASTField(field *ast.Field) reflect.StructTag {
	if field == nil || field.Tag == nil {
		return ""
	}

	return reflect.StructTag(strings.Replace(field.Tag.Value, "`", "", -1))
}

func getImportPkgFromSelectorSel(path string) string {
	path = strings.Replace(path, "\"", "", -1)
	regexp, _ := regexp.Compile(`/v([0-9]+)$`)
	match := regexp.ReplaceAllString(path, "")
	fields := strings.Split(match, "/")

	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

// return abs of import path
// todo:
// should search vendor, go src library ...
func getDirFromImportSpec(value string) string {
	gopath := os.Getenv("GOPATH")
	return gopath + "/src/" + strings.Replace(value, "\"", "", -1)
}

func getImportFromFullPath(path string) (string, string) {
	pkg := filepath.Base(path)
	prefix := os.Getenv("GOPATH") + "/src/"
	if len(prefix) < len(path) {
		return pkg, "\"" + path[len(prefix):] + "\""
	}
	return pkg, ""
}
