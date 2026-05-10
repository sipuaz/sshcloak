package cli

// Version is set at build time via -ldflags:
//
//	-X github.com/sipuaz/sshcloak/cmd/internal/cli.Version=v1.2.3
//
// Local builds fall back to "dev".
var Version = "dev"
