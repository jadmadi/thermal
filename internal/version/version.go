package version

// These are set at build time via ldflags:
//
//	go build -ldflags "-X github.com/jadmadi/thermal/internal/version.Version=v0.2.0"
//
// GoReleaser injects them automatically via the ldflags section in
// .goreleaser.yml. When building from source without ldflags, Version
// defaults to "dev" and Commit to "unknown".

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// String returns a human-readable version string.
func String() string {
	if Version == "dev" {
		return "dev (built from source)"
	}
	return Version
}
