package cli

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/diagnostic"
	"context"
	"fmt"
	"runtime"
	"strings"
)

type app struct {
	ctx     context.Context
	streams Streams
	build   buildinfo.Info
	bundle  bundle.Bundle
	root    string
	json    bool
}

func NewRootCommand(ctx context.Context, streams Streams, build buildinfo.Info) *RootCommand {
	contentBundle, err := bundle.Embedded()
	if err != nil {
		panic(err)
	}
	return &RootCommand{app: &app{ctx: ctx, streams: streams, build: build, bundle: contentBundle}}
}

type RootCommand struct {
	app  *app
	args []string
	path string
}

func (command *RootCommand) SetArgs(args []string) { command.args = args }

func (command *RootCommand) CommandPath() string {
	if command.path == "" {
		return "acctx"
	}
	return "acctx " + command.path
}

func (command *RootCommand) Execute() error {
	original := append([]string(nil), command.args...)
	args := make([]string, 0, len(original))
	for index := 0; index < len(original); index++ {
		switch original[index] {
		case "--json":
			command.app.json = true
		case "--root":
			if index+1 >= len(original) {
				return fmt.Errorf("missing value for --root")
			}
			command.app.root = original[index+1]
			index++
		default:
			args = append(args, original[index])
		}
	}
	if len(args) == 0 {
		return fmt.Errorf("command required")
	}
	command.path = strings.Join(args[:min(2, len(args))], " ")
	switch args[0] {
	case "version":
		return runVersion(command.app, args[1:])
	case "content":
		return runContent(command.app, args[1:])
	case "init":
		return runInit(command.app, args[1:])
	case "project":
		return runProject(command.app, args[1:])
	case "skill":
		return runSkill(command.app, args[1:])
	case "company":
		return runCompany(command.app, args[1:])
	case "year":
		return runYear(command.app, args[1:])
	case "task":
		return runTask(command.app, args[1:])
	case "validate":
		return runValidate(command.app, args[1:])
	case "calc":
		return runCalc(command.app, args[1:])
	default:
		return &ExitError{
			Code: ExitFailure,
			Diagnostic: diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_UNKNOWN_COMMAND",
				Message:  "فرمان ناشناخته است",
			},
		}
	}
}

func (a *app) linuxMutating() error {
	if runtime.GOOS != "linux" {
		return &ExitError{
			Code: ExitConfig,
			Diagnostic: diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_UNSUPPORTED_PLATFORM",
				Message:  "نسخه فعلی acctx فقط Linux را پشتیبانی می‌کند",
			},
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
