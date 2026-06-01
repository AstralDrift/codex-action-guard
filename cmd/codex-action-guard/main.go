package main

import (
	"os"

	"github.com/AstralDrift/codex-action-guard/internal/cli"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr, cli.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}))
}
