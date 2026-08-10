package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/doctor"
	"acctx/internal/gitstate"
	"acctx/internal/output"
	"acctx/internal/plan"
	"acctx/internal/project"
	"acctx/internal/upgrade"
	"fmt"
	"time"
)

func runProject(a *app,args []string)error{if len(args)==0{return fmt.Errorf("project subcommand required")};switch args[0]{case "status":root,e:=resolveProjectRoot(a);if e!=nil{return e};s,e:=project.ReadStatus(root);if e!=nil{return e};return output.Write(a.streams.Out,a.json,output.Result{Command:"project status",Status:"ok",Data:s});case "doctor":root,e:=resolveProjectRoot(a);if e!=nil{return e};r,e:=doctor.Run(root);if e!=nil{return e};if e:=output.Write(a.streams.Out,a.json,output.Result{Command:"project doctor",Status:"ok",Data:r});e!=nil{return e};if !r.Healthy{return &ExitError{Code:ExitValidation,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_DOCTOR_FAILED",Message:"سلامت Workspace تأیید نشد"}}};return nil;case "upgrade":if e:=a.linuxMutating();e!=nil{return e};_,flags,_,e:=parseFlags(args[1:]);if e!=nil{return e};root,e:=resolveProjectRoot(a);if e!=nil{return e};g,e:=gitstate.Inspect(root);if e!=nil{return e};if !g.Clean&&!flags["--allow-dirty"]{return &ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_GIT_WORKTREE_DIRTY",Message:"برای upgrade باید worktree تمیز باشد"}}};p,e:=upgrade.Build(root,a.bundle,time.Now().UTC());if e!=nil{return e};if e:=renderPlan(a,p);e!=nil{return e};if flags["--plan"]{return nil};if !confirm(a,flags["--yes"]){return &ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_CONFIRMATION_REQUIRED",Message:"عملیات تأیید نشد"}}};_,e=plan.Apply(p);if e!=nil{return e};if p.HasConflicts(){return &ExitError{Code:ExitConflict,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_UPGRADE_CONFLICT",Message:"upgrade دارای تعارض است"}}};return nil;default:return fmt.Errorf("unknown project subcommand")}}
