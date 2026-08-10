package plan

import (
	"acctx/internal/fsx"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ApplyResult struct {
	Applied int `json:"applied"`
	Skipped int `json:"skipped"`
	Conflicts int `json:"conflicts"`
}

func Apply(p Plan) (ApplyResult, error) {
	res := ApplyResult{Skipped:p.Summary.Skips, Conflicts:p.Summary.Conflicts}
	stage := filepath.Join(p.Root, ".acctx", "staging", p.ID)
	if err:=os.MkdirAll(stage,0755);err!=nil{return res,err}
	defer os.RemoveAll(stage)
	type backup struct{path string;data []byte;mode os.FileMode;link string;exists bool}
	backs:=[]backup{}
	rollback:=func(){for i:=len(backs)-1;i>=0;i--{b:=backs[i];dst:=filepath.Join(p.Root,filepath.FromSlash(b.path));_ = os.RemoveAll(dst);if b.exists{_ = os.MkdirAll(filepath.Dir(dst),0755);if b.link!=""{_ = os.Symlink(b.link,dst)}else{_ = os.WriteFile(dst,b.data,b.mode)}}}}
	for _,op:=range p.Operations{
		if op.Kind==Skip||op.Kind==Conflict{continue}
		dst,e:=fsx.ResolveInside(p.Root,filepath.FromSlash(op.Path));if e!=nil{rollback();return res,e}
		b:=backup{path:op.Path}
		if fi,e:=os.Lstat(dst);e==nil{b.exists=true;b.mode=fi.Mode();if fi.Mode()&os.ModeSymlink!=0{b.link,_=os.Readlink(dst)}else if fi.Mode().IsRegular(){b.data,_=os.ReadFile(dst)}}
		backs=append(backs,b)
		if op.BeforeDigest!=""&&b.exists&&b.link==""{got,e:=fsx.FileDigest(dst);if e!=nil||got!=op.BeforeDigest{rollback();return res,fmt.Errorf("precondition mismatch %s",op.Path)}}
		switch op.Kind{
		case Create,Update:
			if e:=os.MkdirAll(filepath.Dir(dst),0755);e!=nil{rollback();return res,e};if e:=os.RemoveAll(dst);e!=nil{rollback();return res,e};if e:=os.WriteFile(dst,op.Payload,0644);e!=nil{rollback();return res,e}
		case Link:
			if filepath.IsAbs(op.Target){rollback();return res,fmt.Errorf("absolute symlink target")}
			resolved:=filepath.Clean(filepath.Join(filepath.Dir(dst),op.Target));rel,e:=filepath.Rel(p.Root,resolved);if e!=nil||rel==".."||(len(rel)>3&&rel[:3]=="../"){rollback();return res,fmt.Errorf("link escapes project")}
			if e:=os.MkdirAll(filepath.Dir(dst),0755);e!=nil{rollback();return res,e};_ = os.RemoveAll(dst);if e:=os.Symlink(op.Target,dst);e!=nil{rollback();return res,e}
		case Delete:
			if e:=os.RemoveAll(dst);e!=nil{rollback();return res,e}
		}
		res.Applied++
	}
	if res.Applied>0{
		rec:=struct{SchemaVersion int `yaml:"schema_version"`;OperationID string `yaml:"operation_id"`;Command string `yaml:"command"`;AppliedAt time.Time `yaml:"applied_at"`;Summary Summary `yaml:"summary"`}{1,p.ID,p.Command,time.Now().UTC(),p.Summary}
		b,_:=json.MarshalIndent(rec,"","  ");dir:=filepath.Join(p.Root,".acctx","operations");_ = os.MkdirAll(dir,0755);if e:=os.WriteFile(filepath.Join(dir,p.ID+".yaml"),b,0644);e!=nil{return res,e}
	}
	return res,nil
}
