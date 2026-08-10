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

type app struct{ctx context.Context;streams Streams;build buildinfo.Info;bundle bundle.Bundle;root string;json bool}
func NewRootCommand(ctx context.Context,streams Streams,bi buildinfo.Info)*RootCommand{b,e:=bundle.Embedded();if e!=nil{panic(e)};return &RootCommand{app:&app{ctx:ctx,streams:streams,build:bi,bundle:b}}}
type RootCommand struct{app *app;args []string;path string}
func(c *RootCommand)SetArgs(args []string){c.args=args}
func(c *RootCommand)CommandPath()string{if c.path==""{return "acctx"};return "acctx "+c.path}
func(c *RootCommand)Execute()error{original:=append([]string(nil),c.args...);args:=make([]string,0,len(original));for i:=0;i<len(original);i++{switch original[i]{case "--json":c.app.json=true;case "--root":if i+1>=len(original){return fmt.Errorf("missing value for --root")};c.app.root=original[i+1];i++;default:args=append(args,original[i])}};if len(args)==0{return fmt.Errorf("command required")};c.path=strings.Join(args[:min(2,len(args))]," ");switch args[0]{case "version":return runVersion(c.app,args[1:]);case "content":return runContent(c.app,args[1:]);case "init":return runInit(c.app,args[1:]);case "project":return runProject(c.app,args[1:]);case "skill":return runSkill(c.app,args[1:]);default:return &ExitError{Code:ExitFailure,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_UNKNOWN_COMMAND",Message:"فرمان ناشناخته است"}}}}
func(a *app)linuxMutating()error{if runtime.GOOS!="linux"{return &ExitError{Code:ExitConfig,Diagnostic:diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_UNSUPPORTED_PLATFORM",Message:"نسخه فعلی acctx فقط Linux را پشتیبانی می‌کند"}}};return nil}
func min(a,b int)int{if a<b{return a};return b}
