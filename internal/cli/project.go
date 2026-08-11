package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/doctor"
	"acctx/internal/gitstate"
	"acctx/internal/output"
	"acctx/internal/project"
	"acctx/internal/upgrade"
	"fmt"
	"time"
)

func runProject(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project subcommand required")
	}
	switch args[0] {
	case "status":
		root, err := resolveProjectRoot(a)
		if err != nil {
			return err
		}
		status, err := project.ReadStatus(root)
		if err != nil {
			return err
		}
		return output.Write(a.streams.Out, a.json, output.Result{Command: "project status", Status: "ok", Data: status})
	case "doctor":
		root, err := resolveProjectRoot(a)
		if err != nil {
			return err
		}
		report, err := doctor.Run(root)
		if err != nil {
			return err
		}
		if err := output.Write(a.streams.Out, a.json, output.Result{Command: "project doctor", Status: "ok", Data: report}); err != nil {
			return err
		}
		if !report.Healthy {
			return &ExitError{Code: ExitValidation, Diagnostic: diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_DOCTOR_FAILED", Message: "سلامت Workspace تأیید نشد"}}
		}
		return nil
	case "upgrade":
		if err := a.linuxMutating(); err != nil {
			return err
		}
		_, flags, rest, err := parseFlags(args[1:])
		if err != nil {
			return err
		}
		if len(rest) != 0 {
			return fmt.Errorf("unexpected project upgrade arguments")
		}
		root, err := resolveProjectRoot(a)
		if err != nil {
			return err
		}
		gitInfo, err := gitstate.Inspect(root)
		if err != nil {
			return err
		}
		if !gitInfo.Clean && !flags["--allow-dirty"] {
			return &ExitError{Code: ExitConfig, Diagnostic: diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_GIT_WORKTREE_DIRTY", Message: "برای upgrade باید worktree تمیز باشد"}}
		}
		operationPlan, err := upgrade.Build(root, a.bundle, a.build, time.Now().UTC())
		if err != nil {
			return err
		}
		return applyCommandPlan(a, operationPlan, flags)
	default:
		return fmt.Errorf("unknown project subcommand")
	}
}
