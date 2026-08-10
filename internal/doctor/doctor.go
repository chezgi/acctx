package doctor

import (
	"acctx/internal/diagnostic"
	"acctx/internal/fsx"
	"acctx/internal/gitstate"
	"acctx/internal/managedblock"
	"acctx/internal/manifest"
	"acctx/internal/skills"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Report struct { Healthy bool `json:"healthy"`; Issues []diagnostic.Item `json:"issues"` }

func Run(root string) (Report,error) {
	m,e:=manifest.Load(root); if e!=nil{return Report{},e}
	issues:=[]diagnostic.Item{}
	for _,f:=range m.Managed.Files {
		p:=filepath.Join(root,filepath.FromSlash(f.Path)); d,e:=fsx.FileDigest(p)
		if e!=nil { issues=append(issues,diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_MANAGED_FILE_MISSING",Message:"فایل مدیریت‌شده موجود نیست",Path:f.Path}) } else if d!=f.Digest { issues=append(issues,diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_MANAGED_FILE_DRIFT",Message:"فایل مدیریت‌شده تغییر کرده است",Path:f.Path}) }
	}
	for _,b:=range m.Managed.Blocks {
		p:=filepath.Join(root,b.Path); raw,e:=os.ReadFile(p)
		if e!=nil { issues=append(issues,diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_MANAGED_BLOCK_MISSING",Message:"فایل block موجود نیست",Path:b.Path}); continue }
		style:=managedblock.Markdown; if b.Style=="gitignore" { style=managedblock.Gitignore }
		body,e:=managedblock.Body(raw,style)
		if e!=nil { issues=append(issues,diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_MANAGED_BLOCK_INVALID",Message:"نشانه‌های block معتبر نیستند",Path:b.Path}) } else if fsx.BytesDigest(body)!=b.BodyDigest { issues=append(issues,diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_MANAGED_BLOCK_DRIFT",Message:"محتوای block تغییر کرده است",Path:b.Path}) }
	}
	svc:=skills.New(root); for _,s:=range m.Managed.Skills { st,e:=svc.Status(s.ID); if e!=nil||!st.Valid { issues=append(issues,diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_SKILL_INVALID",Message:"Skill یا link آن معتبر نیست",Path:s.ID}) } }
	if ds,e:=os.ReadDir(filepath.Join(root,".acctx","staging")); e==nil&&len(ds)>0 { issues=append(issues,diagnostic.Item{Severity:diagnostic.Warning,Code:"ACCTX_STAGING_LEFTOVER",Message:"فایل موقت staging باقی مانده است",Path:".acctx/staging"}) }
	if _,e:=os.Stat(filepath.Join(root,".acctx","recovery"));e==nil { issues=append(issues,diagnostic.Item{Severity:diagnostic.Error,Code:"ACCTX_FORBIDDEN_RECOVERY_DIRECTORY",Message:"پوشه recovery نباید وجود داشته باشد",Path:".acctx/recovery"}) }
	if g,e:=gitstate.Inspect(root);e==nil&&!g.Clean { issues=append(issues,diagnostic.Item{Severity:diagnostic.Warning,Code:"ACCTX_GIT_WORKTREE_DIRTY",Message:"Git worktree دارای تغییرات ثبت‌نشده است",Path:strings.Join(g.ChangedPaths,",")}) }
	sort.Slice(issues,func(i,j int)bool{if issues[i].Severity==issues[j].Severity{if issues[i].Code==issues[j].Code{return issues[i].Path<issues[j].Path};return issues[i].Code<issues[j].Code};return issues[i].Severity>issues[j].Severity})
	healthy:=true;for _,i:=range issues{if i.Severity==diagnostic.Error{healthy=false}}
	return Report{healthy,issues},nil
}
