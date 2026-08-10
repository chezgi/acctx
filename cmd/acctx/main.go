package main

import (
	"acctx/internal/buildinfo"
	"acctx/internal/cli"
	"context"
	"os"
)

func main() {
	os.Exit(cli.Execute(context.Background(), os.Args[1:], cli.DefaultStreams(), buildinfo.Current()))
}
