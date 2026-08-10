package cli

import (
	"acctx/internal/buildinfo"
	"acctx/internal/diagnostic"
	"acctx/internal/output"
	"context"
	"errors"
	"io"
	"os"
)

const (
	ExitOK = 0
	ExitFailure = 1
	ExitValidation = 2
	ExitConflict = 3
	ExitVersion = 4
	ExitConfig = 5
)

type Streams struct { In io.Reader; Out io.Writer; Err io.Writer; Interactive bool }
type ExitError struct { Code int; Diagnostic diagnostic.Item }
func (e *ExitError) Error() string { return e.Diagnostic.Message }
func Execute(ctx context.Context, args []string, streams Streams, bi buildinfo.Info) int {
	cmd:=NewRootCommand(ctx,streams,bi);cmd.SetArgs(args)
	if e:=cmd.Execute();e!=nil{
		var x *ExitError
		if errors.As(e,&x){_ = output.Write(streams.Err,isJSON(args),output.Result{Command:cmd.CommandPath(),Status:"error",Diagnostics:[]diagnostic.Item{x.Diagnostic}});return x.Code}
		_ = output.Write(streams.Err,isJSON(args),output.Result{Command:cmd.CommandPath(),Status:"error",Diagnostics:[]diagnostic.Item{{Severity:diagnostic.Error,Code:"ACCTX_COMMAND_FAILED",Message:e.Error()}}});return ExitFailure
	}
	return ExitOK
}
func isJSON(args []string)bool{for _,a:=range args{if a=="--json"{return true}};return false}
func DefaultStreams() Streams { return Streams{In:os.Stdin,Out:os.Stdout,Err:os.Stderr,Interactive:true} }
