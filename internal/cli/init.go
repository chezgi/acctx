package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/output"
	"acctx/internal/plan"
	"acctx/internal/project"
	"acctx/internal/workspace"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type bootstrap struct { SchemaVersion int `json:"schema_version"`; ProjectID string `json:"project_id"`; Preset string `json:"preset"` }

func resolveGitRoot(a *app)(string,error){start:=a.root;if start==""{start,_=os.Getwd()};r,e:=project.GitRoot(start);if e!=nil{return "",&ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_GIT_REPOSITORY_NOT_FOUND",Message:"Git repository پیدا نشد"}}};return r,nil}
func resolveProjectRoot(a *app)(string,error){start:=a.root;if start==""{start,_=os.Getwd()};r,e:=project.Root(start);if e!=nil{return "",&ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_PROJECT_NOT_FOUND",Message:"پروژه acctx پیدا نشد"}}};return r,nil}
func renderPlan(a *app,p plan.Plan)error{if a.json{return output.Write(a.streams.Out,true,output.Result{Command:p.Command,Status:"ok",Data:p})};for _,o:=range p.Operations{fmt.Fprintf(a.streams.Out,"%-8s %s",o.Kind,o.Path);if o.Target!=""{fmt.Fprintf(a.streams.Out," -> %s",o.Target)};if o.Message!=""{fmt.Fprintf(a.streams.Out," (%s)",o.Message)};fmt.Fprintln(a.streams.Out)};return nil}
func confirm(a *app,yes bool)bool{if yes{return true};if !a.streams.Interactive{return false};fmt.Fprint(a.streams.Out,"اعمال شود؟ [y/N] ");s,_:=bufio.NewReader(a.streams.In).ReadString('\n');s=strings.TrimSpace(strings.ToLower(s));return s=="y"||s=="yes"}
func parseFlags(args []string)(map[string]string,map[string]bool,[]string,error){vals:=map[string]string{};bools:=map[string]bool{};rest:=[]string{};for i:=0;i<len(args);i++{a:=args[i];switch a{case "--plan","--yes","--non-interactive","--allow-dirty","--force":bools[a]=true;default:if strings.HasPrefix(a,"--"){if i+1>=len(args){return nil,nil,nil,fmt.Errorf("missing value for %s",a)};vals[a]=args[i+1];i++}else{rest=append(rest,a)}}};return vals,bools,rest,nil}
func runInit(a *app,args []string)error{if e:=a.linuxMutating();e!=nil{return e};vals,flags,_,e:=parseFlags(args);if e!=nil{return e};root,e:=resolveGitRoot(a);if e!=nil{return e};projectID:=vals["--project-id"];preset:=vals["--preset"];if config:=vals["--config"];config!=""{b,e:=os.ReadFile(config);if e!=nil{return e};var x bootstrap;if e:=json.Unmarshal(b,&x);e!=nil{return fmt.Errorf("bootstrap config must be JSON-compatible YAML: %w",e)};if x.SchemaVersion!=0&&x.SchemaVersion!=1{return &ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_CONFIG_INVALID",Message:"config schema معتبر نیست"}}};if projectID==""{projectID=x.ProjectID};if preset==""{preset=x.Preset}};if projectID==""{projectID=filepath.Base(root)};p,_,e:=workspace.BuildInitPlan(root,workspace.InitOptions{ProjectID:projectID,Preset:preset},a.bundle,a.build,time.Now().UTC());if e!=nil{return e};if e:=renderPlan(a,p);e!=nil{return e};if flags["--plan"]{return nil};interactive:=a.streams.Interactive&&!flags["--non-interactive"];if !interactive&&!flags["--yes"]{return &ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_CONFIRMATION_REQUIRED",Message:"برای اجرای غیرتعاملی --yes لازم است"}}};old:=a.streams.Interactive;a.streams.Interactive=interactive;ok:=confirm(a,flags["--yes"]);a.streams.Interactive=old;if !ok{return &ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_CONFIRMATION_REQUIRED",Message:"عملیات تأیید نشد"}}};_,e=plan.Apply(p);if e!=nil{return e};if p.HasConflicts(){return &ExitError{Code:ExitConflict,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_INIT_PARTIAL_CONFLICT",Message:"بخش‌های بدون تعارض اعمال شدند؛ تعارض‌ها باقی مانده‌اند"}}};return nil}
