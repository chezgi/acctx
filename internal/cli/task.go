package cli

import (
	"acctx/internal/output"
	"acctx/internal/task"
	"fmt"
	"time"
)

func runTask(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task subcommand required")
	}
	root, err := resolveProjectRoot(a)
	if err != nil {
		return err
	}
	switch args[0] {
	case "types":
		return output.Write(a.streams.Out, a.json, output.Result{Command: "task types", Status: "ok", Data: task.Definitions()})
	case "init":
		if err := a.linuxMutating(); err != nil {
			return err
		}
		values, flags, rest, err := parseFlags(args[1:])
		if err != nil {
			return err
		}
		if len(rest) != 1 {
			return fmt.Errorf("usage: acctx task init <type> --year <id>")
		}
		operationPlan, _, err := task.BuildInitPlan(root, rest[0], task.Options{
			Year:   values["--year"],
			Period: values["--period"],
		}, a.bundle, time.Now().UTC())
		if err != nil {
			return err
		}
		return applyCommandPlan(a, operationPlan, flags)
	case "list":
		values, _, _, err := parseFlags(args[1:])
		if err != nil {
			return err
		}
		if values["--year"] == "" {
			return fmt.Errorf("--year is required")
		}
		models, err := task.List(root, values["--year"])
		if err != nil {
			return err
		}
		return output.Write(a.streams.Out, a.json, output.Result{Command: "task list", Status: "ok", Data: models})
	case "status":
		values, _, rest, err := parseFlags(args[1:])
		if err != nil {
			return err
		}
		if values["--year"] == "" || len(rest) != 1 {
			return fmt.Errorf("usage: acctx task status <id> --year <year>")
		}
		status, err := task.ReadStatus(root, values["--year"], rest[0])
		if err != nil {
			return err
		}
		return output.Write(a.streams.Out, a.json, output.Result{Command: "task status", Status: "ok", Data: status})
	default:
		return fmt.Errorf("unknown task subcommand")
	}
}
