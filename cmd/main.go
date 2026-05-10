package main

import (
	"os"

	"github.com/sipuaz/sshcloak/cmd/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
