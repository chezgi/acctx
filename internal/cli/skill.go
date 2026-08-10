package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/output"
	"acctx/internal/plan"
	"acctx/internal/skills"
	"fmt"
	"time"
)

func runSkill(a *app,args []string)error{if len(args)==0{return fmt.Errorf("skill subcommand required")};root,e:=resolveProjectRoot(a);if e!=nil{return e};svc:=skills.New(root);switch args[0]{case "list":v,e:=svc.List();if e!=nil{return e};return output.Write(a.streams.Out,a.json,output.Result{Command:"skill list",Status:"ok",Data:v});case "status","validate":if len(args)<2{return fmt.Errorf("skill id required")};v,e:=svc.Status(args[1]);if e!=nil{return e};if e:=output.Write(a.streams.Out,a.json,output.Result{Command:"skill "+args[0],Status:"ok",Data:v});e!=nil{return e};if args[0]=="validate"&&!v.Valid{return &ExitError{Code:ExitValidation,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_SKILL_INVALID",Message:"Skill معتبر نیست"}}};return nil;case "diff":if len(args)<2{return fmt.Errorf("skill id required")};d,e:=svc.Diff(args[1]);if e!=nil{return e};if a.json{return output.Write(a.streams.Out,true,output.Result{Command:"skill diff",Status:"ok",Data:map[string]string{"diff":d}})};fmt.Fprint(a.streams.Out,d);return nil;case "override","reset","adopt":if len(args)<2{return fmt.Errorf("skill id required")};if e:=a.linuxMutating();e!=nil{return e};_,flags,_,e:=parseFlags(args[2:]);if e!=nil{return e};var p plan.Plan;switch args[0]{case "override":p,e=svc.OverridePlan(args[1],time.Now().UTC());case "reset":p,e=svc.ResetPlan(args[1],time.Now().UTC(),flags["--force"]);case "adopt":p,e=svc.AdoptPlan(args[1],time.Now().UTC())};if e!=nil{return e};if e:=renderPlan(a,p);e!=nil{return e};if flags["--plan"]{return nil};if !confirm(a,flags["--yes"]){return &ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_CONFIRMATION_REQUIRED",Message:"عملیات تأیید نشد"}}};_,e=plan.Apply(p);if e!=nil{return e};if p.HasConflicts(){return &ExitError{Code:ExitConflict,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_SKILL_CONFLICT",Message:"عملیات Skill دارای تعارض است"}}};return nil;default:return fmt.Errorf("unknown skill subcommand")}}
