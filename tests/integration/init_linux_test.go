//go:build linux

package integration_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"acctx/internal/buildinfo"
	"acctx/internal/cli"
)

func TestInitCreatesRelativeSkillLinks(t *testing.T){root:=t.TempDir();if err:=os.Mkdir(filepath.Join(root,".git"),0o755);err!=nil{t.Fatal(err)};if err:=os.WriteFile(filepath.Join(root,".git","HEAD"),[]byte("ref: refs/heads/main\n"),0o644);err!=nil{t.Fatal(err)};var out,errb bytes.Buffer;c:=cli.Execute(context.Background(),[]string{"--root",root,"init","--non-interactive","--yes"},cli.Streams{In:bytes.NewBuffer(nil),Out:&out,Err:&errb,Interactive:false},buildinfo.Info{Version:"test"});if c!=0{t.Fatalf("code=%d out=%s err=%s",c,out.String(),errb.String())};for _,p:=range []string{".claude/skills/acctx-workspace",".agents/skills/acctx-source-review"}{target,e:=os.Readlink(filepath.Join(root,filepath.FromSlash(p)));if e!=nil{t.Fatal(e)};if filepath.IsAbs(target){t.Fatalf("absolute link %s",target)}}}
