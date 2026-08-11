package cli

import (
	"acctx/internal/output"
	"acctx/internal/year"
	"fmt"
	"strconv"
	"time"
)

func runYear(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("year subcommand required")
	}
	root, err := resolveProjectRoot(a)
	if err != nil {
		return err
	}
	switch args[0] {
	case "init":
		if err := a.linuxMutating(); err != nil {
			return err
		}
		values, flags, rest, err := parseFlags(args[1:])
		if err != nil {
			return err
		}
		if len(rest) != 1 {
			return fmt.Errorf("usage: acctx year init <id> [--mode mode]")
		}
		mode := values["--mode"]
		if flags["--historical"] {
			mode = year.ModeHistorical
		}
		if flags["--archive"] {
			mode = year.ModeArchive
		}
		rulesetYear := 0
		if raw := values["--ruleset-year"]; raw != "" {
			rulesetYear, err = strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("invalid --ruleset-year: %w", err)
			}
		}
		operationPlan, _, err := year.BuildInitPlan(root, rest[0], year.Options{
			Mode:        mode,
			StartsOn:    values["--start"],
			EndsOn:      values["--end"],
			RulesetYear: rulesetYear,
		}, time.Now().UTC())
		if err != nil {
			return err
		}
		return applyCommandPlan(a, operationPlan, flags)
	case "list":
		models, err := year.List(root)
		if err != nil {
			return err
		}
		return output.Write(a.streams.Out, a.json, output.Result{Command: "year list", Status: "ok", Data: models})
	case "status":
		if len(args) != 2 {
			return fmt.Errorf("usage: acctx year status <id>")
		}
		status, err := year.ReadStatus(root, args[1])
		if err != nil {
			return err
		}
		return output.Write(a.streams.Out, a.json, output.Result{Command: "year status", Status: "ok", Data: status})
	default:
		return fmt.Errorf("unknown year subcommand")
	}
}
