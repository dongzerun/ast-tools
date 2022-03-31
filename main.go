package main

import (
	"time"

	"net/http"

	pkgcmd "github.com/dongzerun/ast-tools/cmd"
)

// Manager ...
type Manager struct {
	// Same      string
	// All       bool   `json:"all"`
	// BuildTags string `mapstructure:"tags" json:"build-tags"`
	// Version   int    `json:"-"`
	// TXXXXXX   Test
	// ByteField byte
	// Test
	// Test1         pkgcmd.Test
	// NormalStruct  pkgcmd.RootApp
	// PointerStruct *pkgcmd.RootApp

	// pkgcmd.RootApp
	// RootAPPXXXXXX pkgcmd.RootApp2
	// Source        pkgcmd.Source

	// PointInt    *int
	// PointString *string `json:"pointstring"`
	// PointerSliceInt    *[]int    `json:"pointer-slice-int"`
	// PointerSlicestring *[]string `json:"pointer-slice-string"`
	// PointerSlicePointerInt *[]*int `json:"pointer-slice-pointer-int"`
	// PointerArrayInt *[8]int `json:"pointer-array-int"`
	// PointerArrayPointerInt *[8]*int `json:"pointer-array-pointer-int"`
	// PointerSliceRootApp *[]*pkgcmd.RootApp
	// PointerMapInt   *map[string]int
	// PointerMap   *map[string]*pkgcmd.RootApp

	// Mocker
	// mocker Mocker
	// Err    error

	// innerSource Source

	// Slices       []int
	// SliceBytes []byte
	// SlicesString []string
	// Arrayint     [8]int
	// ArrayString  [10]string
	// AnotherSlice []int `json:"another_slice"`

	// SliceIntPointer []*int
	// ArrayPointer    [8]*int
	// SliceIntPointer2    []*int
	// SliceStringPointer  []*string
	// ArrayStruct         [8]pkgcmd.RootApp
	// ArrayStructPointer  [8]*pkgcmd.RootApp
	// SliceStruct        []pkgcmd.RootApp
	// SliceStructPointer []*pkgcmd.RootApp
	// SlicesStruct        []pkgcmd.Source
	// SlicesPointerStruct []*pkgcmd.Source
	// SlicesTypeList []Test

	// Maps            map[string]string `json:"another_maps"`
	// Maps1           map[string]string
	// MapsInt         map[string]int
	// MapsStarString  map[string]*string
	// MapsStarString1 map[string]*string
	// MapStarInt1       map[string]*int
	// MapsStruct map[string]pkgcmd.RootApp2
	// MapsPointerStruct map[string]*pkgcmd.RootApp2
	// MapsAlias         map[string]pkgcmd.Source
	// MapsPointerAlias  map[string]*pkgcmd.Source

	// MapsAliasKey map[pkgcmd.Source]string
	// MyList PackageList

	// // go src library
	// createAt        time.Time      `json:"Creat_AT"`
	// pointerTime     *time.Time     `json:"pointer_time"`
	// last            time.Duration  `json:"wocao"`
	// pointerDuration *time.Duration `json:"pointer_duration"`
	// request         http.Request   `json:"http_request"`
	// pointerRequest  *http.Request  `json:"pointer_request"`
	// err             error          `json:"INNER_ERROR"`

	// MapTime        map[string]time.Time
	// SliceTime      []time.Time
	// ArrayTime      [10086]time.Time
	// MapPointerTime map[string]*time.Time
	// grabtime       grabtime.RFC3339
	// st             stateNode
	// P Package
	// Pp    *Package
	// AnCfg *AnotherCfg
}

// PackageList array of Package
type PackageList []*Package

// PackageMap ...
type PackageMap map[string]string

// AnotherCfg ...
type AnotherCfg struct {
	Address string
	Port    int
}

// Package wraps the array of Package
type Package struct {
	Name      string `json:"name"`
	Corrupted bool   `json:"-"`
	Create    time.Time
}

type stateNode struct {
	name string
	next *stateNode
}

//   &main.Manager{
//   	Same:      "",
//   	All:       false,
// - 	BuildTags: "aaaa",
// + 	BuildTags: "",
//   	P: main.Package{
// - 		Name:      "",
// + 		Name:      "wocao",
//   		Corrupted: false,
// - 		Create:    s"0001-01-01 00:00:00 +0000 UTC",
// + 		Create:    s"2022-03-14 13:31:56.604877 +0800 CST m=+0.657110754",
//   	},
// - 	AnCfg: &main.AnotherCfg{},
// + 	AnCfg: nil,
//   }

// User ...
type User struct {
	UserName string
}

func main() {
	var _ http.Request
	pkgcmd.Execute()
	// m1 := &Manager{
	// 	BuildTags: "aaaa",
	// 	AnCfg:     &AnotherCfg{},
	// 	Maps1:     map[string]string{"a": "b", "c": "d"},
	// }
	// m2 := &Manager{
	// 	P:      Package{Name: "wocao", Create: time.Now()},
	// 	Mocker: &FakeMocker{MockName: "dzr"},
	// }
	// diff := cmp.Diff(m1, m2, cmpopts.IgnoreUnexported())
	// fmt.Println(diff)
	// mcopy := deepcopy.Copy(m2)
	// fmt.Println(cmp.Diff(m1, mcopy, cmpopts.IgnoreUnexported()))

}
