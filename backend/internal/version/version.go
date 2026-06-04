package version

// BuildInfo describes the running backend build. The fields are intentionally
// simple strings so container builds can inject them through -ldflags later.
type BuildInfo struct {
	Service string `json:"service"`
	Version string `json:"version"`
	Commit  string `json:"commit,omitempty"`
}

var (
	Version = "mvp-7"
	Commit  = "dev"
)

func Current() BuildInfo {
	return BuildInfo{
		Service: "logarift-api",
		Version: Version,
		Commit:  Commit,
	}
}
