package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"acctx/internal/buildinfo"
	"acctx/internal/cli"
	"acctx/internal/output"
)

func fakeGitRepo(t *testing.T) string { t.Helper(); root:=t.TempDir(); if err:=os.Mkdir(filepath.Join(root,".git"),0o755);err!=nil{t.Fatal(err)}; if err:=os.WriteFile(filepath.Join(root,".git","HEAD"),[]byte("ref: refs/heads/main\n"),0o644);err!=nil{t.Fatal(err)}; return root }
func TestExecuteVersionJSON(t *testing.T){var out,errb bytes.Buffer;code:=cli.Execute(context.Background(),[]string{"--json","version"},cli.Streams{In:bytes.NewBuffer(nil),Out:&out,Err:&errb,Interactive:false},buildinfo.Info{Version:"0.1.0",Commit:"abc"});if code!=0{t.Fatalf("code=%d err=%s",code,errb.String())};var r output.Result;if e:=json.Unmarshal(out.Bytes(),&r);e!=nil{t.Fatal(e)};if r.Status!="ok"||r.Command!="version"{t.Fatalf("%#v",r)}}
func TestInitPlanAndApply(t *testing.T){root:=fakeGitRepo(t);var out,errb bytes.Buffer;run:=func(args ...string)int{out.Reset();errb.Reset();return cli.Execute(context.Background(),append([]string{"--root",root},args...),cli.Streams{In:bytes.NewBuffer(nil),Out:&out,Err:&errb,Interactive:false},buildinfo.Info{Version:"0.1.0"})};if c:=run("init","--plan","--json");c!=0{t.Fatalf("plan %d %s",c,errb.String())};if _,e:=os.Stat(filepath.Join(root,".acctx","manifest.yaml"));!os.IsNotExist(e){t.Fatal("plan mutated filesystem")};if c:=run("init","--non-interactive","--yes","--json");c!=0{t.Fatalf("apply %d out=%s err=%s",c,out.String(),errb.String())};if _,e:=os.Stat(filepath.Join(root,".acctx","manifest.yaml"));e!=nil{t.Fatal(e)};if c:=run("project","doctor","--json");c!=0{t.Fatalf("doctor %d out=%s err=%s",c,out.String(),errb.String())}}
