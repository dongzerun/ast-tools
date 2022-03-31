package config

import "runtime/debug"

const (
	_defaultSemVer = "v0.0.0-dev"
)

// SemVer is the version of mockery at build time.
var SemVer = ""

// GetSemverInfo ...
func GetSemverInfo() string {
	if SemVer != "" {
		return SemVer
	}
	version, ok := debug.ReadBuildInfo()
	if ok && version.Main.Version != "(devel)" {
		return version.Main.Version
	}
	return _defaultSemVer
}

// Version ...
type Version struct {
	GA int
}

// Config ...
type Config struct {
	Action       string
	IgnoreFields []string `mapstructure:"ignorefields"`

	All                  bool
	BuildTags            string `mapstructure:"tags"`
	Case                 string
	Config               string
	Cpuprofile           string
	Dir                  string
	DisableVersionString bool `mapstructure:"disable-version-string"`
	DryRun               bool `mapstructure:"dry-run"`
	Exported             bool `mapstructure:"exported"`
	FileName             string
	InPackage            bool
	KeepTree             bool
	LogLevel             string `mapstructure:"log-level"`
	Name                 string
	Outpkg               string
	Packageprefix        string
	Output               string
	Print                bool
	Quiet                bool
	Recursive            bool
	Source               string `mapstructure:"source"`
	SrcPkg               string `mapstructure:"src-pkg"`
	SrcTag               string `mapstructure:"srctag"`
	Target               string `mapstructure:"target"`
	TargetPkg            string `mapstructure:"target-pkg"`
	TargetTag            string `mapstructure:"targettag"`
	BoilerplateFile      string `mapstructure:"boilerplate-file"`

	StructName     string
	Tags           string
	TestOnly       bool
	UnrollVariadic bool `mapstructure:"unroll-variadic"`
	Version        bool
}
