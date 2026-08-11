package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/plan"
)

func applyCommandPlan(a *app, operationPlan plan.Plan, flags map[string]bool) error {
	if err := renderPlan(a, operationPlan); err != nil {
		return err
	}
	if flags["--plan"] {
		return nil
	}
	interactive := a.streams.Interactive && !flags["--non-interactive"]
	if !interactive && !flags["--yes"] {
		return &ExitError{
			Code: ExitConfig,
			Diagnostic: diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_CONFIRMATION_REQUIRED",
				Message:  "برای اجرای غیرتعاملی --yes لازم است",
			},
		}
	}
	previousInteractive := a.streams.Interactive
	a.streams.Interactive = interactive
	confirmed := confirm(a, flags["--yes"])
	a.streams.Interactive = previousInteractive
	if !confirmed {
		return &ExitError{
			Code: ExitConfig,
			Diagnostic: diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_CONFIRMATION_REQUIRED",
				Message:  "عملیات تأیید نشد",
			},
		}
	}
	if _, err := plan.Apply(operationPlan); err != nil {
		return err
	}
	if operationPlan.HasConflicts() {
		return &ExitError{
			Code: ExitConflict,
			Diagnostic: diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_OPERATION_CONFLICT",
				Message:  "بخش‌های بدون تعارض اعمال شدند؛ تعارض‌ها باقی مانده‌اند",
			},
		}
	}
	return nil
}
