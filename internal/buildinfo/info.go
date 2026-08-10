package buildinfo

var Version = "dev"
var Commit = "unknown"
var BuiltAt = "unknown"

type Info struct {
	Version string `json:"version" yaml:"version"`
	Commit string `json:"commit" yaml:"commit"`
	BuiltAt string `json:"built_at" yaml:"built_at"`
}

func Current() Info { return Info{Version: Version, Commit: Commit, BuiltAt: BuiltAt} }
