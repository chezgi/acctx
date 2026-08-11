package cli

import (
	"acctx/internal/evidence"
	"acctx/internal/output"
	"fmt"
	"time"
)

func runEvidence(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("evidence subcommand required")
	}
	if args[0] != "index" {
		return fmt.Errorf("unknown evidence subcommand")
	}
	root, err := resolveProjectRoot(a)
	if err != nil {
		return err
	}
	values, flags, rest, err := parseFlags(args[1:])
	if err != nil {
		return err
	}
	if len(rest) != 0 || values["--year"] == "" || values["--task"] == "" {
		return fmt.Errorf("usage: acctx evidence index --year <year> --task <task-id> [--output <path>]")
	}
	exclude := []string{}
	if values["--output"] != "" {
		exclude = append(exclude, values["--output"])
	}
	index, err := evidence.BuildTask(root, values["--year"], values["--task"], evidence.Options{
		IncludeCompany:    flags["--include-company"],
		IncludeYearInputs: flags["--include-year-inputs"],
		ExcludePaths:      exclude,
	}, time.Now().UTC())
	if err != nil {
		return err
	}
	if values["--output"] != "" {
		if err := writeJSONOutput(root, values["--output"], index); err != nil {
			return err
		}
	}
	return output.Write(a.streams.Out, a.json, output.Result{Command: "evidence index", Status: "ok", Data: index})
}
