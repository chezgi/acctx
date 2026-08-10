package workspace

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/fsx"
	"acctx/internal/managedblock"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type InitOptions struct { ProjectID string; Preset string }

func readMaybe(path string) ([]byte,error){b,e:=os.ReadFile(path);if os.IsNotExist(e){return nil,nil};return b,e}
func managedFileOp(root,path string,data []byte,prev map[string]manifest.File) plan.Operation {
	dst:=filepath.Join(root,filepath.FromSlash(path));fi,e:=os.Lstat(dst);after:=fsx.BytesDigest(data)
	if os.IsNotExist(e){return plan.Operation{Kind:plan.Create,Path:path,AfterDigest:after,Payload:data}}
	if e!=nil||!fi.Mode().IsRegular(){return plan.Operation{Kind:plan.Conflict,Path:path,Message:"مسیر موجود قابل مدیریت نیست"}}
	got,_:=fsx.FileDigest(dst);if got==after{return plan.Operation{Kind:plan.Skip,Path:path,AfterDigest:after}}
	if p,ok:=prev[path];ok&&got==p.Digest{return plan.Operation{Kind:plan.Update,Path:path,BeforeDigest:got,AfterDigest:after,Payload:data}}
	return plan.Operation{Kind:plan.Conflict,Path:path,Message:"فایل موجود تحت مدیریت acctx نیست یا تغییر کرده است"}
}

func BuildInitPlan(root string,opts InitOptions,b bundle.Bundle,bi buildinfo.Info,now time.Time)(plan.Plan,manifest.Model,error){
	if opts.ProjectID==""{opts.ProjectID=filepath.Base(root)};if opts.Preset==""{opts.Preset=b.Catalog.DefaultPreset};if opts.Preset!=b.Catalog.DefaultPreset{return plan.Plan{},manifest.Model{},fmt.Errorf("ACCTX_PRESET_NOT_FOUND: %s",opts.Preset)}
	var old *manifest.Model;if m,e:=manifest.Load(root);e==nil{old=&m;if m.Project.ID!=opts.ProjectID||m.Project.Preset!=opts.Preset{return plan.Plan{},manifest.Model{},fmt.Errorf("project configuration conflict")}}
	prevFiles:=map[string]manifest.File{};prevSkills:=map[string]manifest.Skill{};if old!=nil{for _,x:=range old.Managed.Files{prevFiles[x.Path]=x};for _,x:=range old.Managed.Skills{prevSkills[x.ID]=x}}
	ops:=[]plan.Operation{};m:=manifest.Model{SchemaVersion:1,Project:manifest.Project{ID:opts.ProjectID,Preset:opts.Preset,InitializedAt:now.UTC()},Generator:manifest.Generator{CLIVersion:bi.Version,ContentVersion:b.Catalog.ContentVersion,ContentDigest:fsx.BytesDigest([]byte(b.Catalog.ContentVersion))}};if old!=nil{m.Project.InitializedAt=old.Project.InitializedAt}
	blockDefs:=[]struct{path string;body []byte;style managedblock.Style}{{"AGENTS.md",[]byte("## acctx project\n\n- Manifest: `.acctx/manifest.yaml`\n- Company: `accounting/company/`\n- Fiscal years: `accounting/fiscal-years/`\n- Skills: `skills/`\n- Read the relevant skill before work.\n- Use acctx for deterministic operations only.\n"),managedblock.Markdown},{"CLAUDE.md",[]byte("@AGENTS.md\n\nClaude Code skills are available under `.claude/skills/`.\n"),managedblock.Markdown},{".gitignore",[]byte(".acctx/cache/\n.acctx/staging/\n.acctx/tmp/\n"),managedblock.Gitignore}}
	for _,d:=range blockDefs{existing,e:=readMaybe(filepath.Join(root,d.path));if e!=nil{return plan.Plan{},m,e};merged,e:=managedblock.Merge(existing,d.body,d.style);if e!=nil{ops=append(ops,plan.Operation{Kind:plan.Conflict,Path:d.path,Message:e.Error()});continue};kind:=plan.Skip;if len(existing)==0{kind=plan.Create}else if merged.Changed{kind=plan.Update};ops=append(ops,plan.Operation{Kind:kind,Path:d.path,BeforeDigest:fsx.BytesDigest(existing),AfterDigest:fsx.BytesDigest(merged.Content),Payload:merged.Content});styleName:="markdown";if d.path==".gitignore"{styleName="gitignore"};m.Managed.Blocks=append(m.Managed.Blocks,manifest.Block{Path:d.path,BodyDigest:fsx.BytesDigest(bytes.Trim(d.body,"\r\n")),Style:styleName})}
	company:=map[string]string{"accounting/company/identity.yaml":"schema_version: 1\nlegal_name_fa: \"\"\nlegal_name_en: \"\"\nnational_id: \"\"\nregistration_number: \"\"\nlegal_type: \"\"\n","accounting/company/registrations.yaml":"schema_version: 1\neconomic_number: \"\"\ntaxpayer_unit_code: \"\"\nregistration_date_jalali: \"\"\n","accounting/company/tax.yaml":"schema_version: 1\nfiscal_year_default: \"\"\nvat_registered: null\ntax_office_name_fa: \"\"\n","accounting/company/contacts.yaml":"schema_version: 1\nregistered_address_fa: \"\"\npostal_code: \"\"\nphone: \"\"\nemail: \"\"\n","accounting/company/bank-accounts.yaml":"schema_version: 1\naccounts: []\n","accounting/company/knowledge-based.yaml":"schema_version: 1\nstatus: unknown\napprovals: []\n","accounting/company/technology-park.yaml":"schema_version: 1\nmember: true\npark_name_fa: \"\"\napprovals: []\n"}
	for p,s:=range company{op:=managedFileOp(root,p,[]byte(s),prevFiles);ops=append(ops,op);if op.Kind!=plan.Conflict{m.Managed.Files=append(m.Managed.Files,manifest.File{Path:p,Digest:fsx.BytesDigest([]byte(s)),SourceID:"bootstrap:"+p})}}
	for _,p:=range []string{"accounting/fiscal-years/.gitkeep","skills/company/.gitkeep","templates/company/.gitkeep","workflows/company/.gitkeep","references/company/.gitkeep","examples/.gitkeep"}{op:=managedFileOp(root,p,[]byte{},prevFiles);ops=append(ops,op);if op.Kind!=plan.Conflict{m.Managed.Files=append(m.Managed.Files,manifest.File{Path:p,Digest:fsx.BytesDigest(nil),SourceID:"workspace-placeholder"})}}
	for _,sk:=range b.Catalog.Skills{files,e:=b.ReadTree(sk.Path);if e!=nil{return plan.Plan{},m,e};vendor:="skills/vendor/"+sk.ID;for rel,data:=range files{p:=filepath.ToSlash(filepath.Join(vendor,rel));op:=managedFileOp(root,p,data,prevFiles);ops=append(ops,op);if op.Kind!=plan.Conflict{m.Managed.Files=append(m.Managed.Files,manifest.File{Path:p,Digest:fsx.BytesDigest(data),SourceID:"skill:"+sk.ID})}};active:=vendor;links:=map[string]string{};if oldsk,ok:=prevSkills[sk.ID];ok&&oldsk.Override!=nil{active=oldsk.ActivePath};for _,provider:=range []string{".claude/skills",".agents/skills"}{lp:=filepath.ToSlash(filepath.Join(provider,sk.ID));target:="../../"+active;dst:=filepath.Join(root,filepath.FromSlash(lp));fi,e:=os.Lstat(dst);kind:=plan.Link;if e==nil{if fi.Mode()&os.ModeSymlink!=0{got,_:=os.Readlink(dst);if got==target{kind=plan.Skip}else{kind=plan.Conflict}}else{kind=plan.Conflict}}else if !os.IsNotExist(e){return plan.Plan{},m,e};ops=append(ops,plan.Operation{Kind:kind,Path:lp,Target:target,Message:"provider path conflict"});if kind!=plan.Conflict{links[provider]=target}};tree,_:=b.ReadTree(sk.Path);m.Managed.Skills=append(m.Managed.Skills,manifest.Skill{ID:sk.ID,Version:sk.Version,VendorPath:vendor,ActivePath:active,Digest:bundle.TreeDigest(tree),ProviderLinks:links})}
	mb,e:=manifest.Marshal(m);if e!=nil{return plan.Plan{},m,e};mp:=".acctx/manifest.yaml";existing,_:=readMaybe(filepath.Join(root,mp));mk:=plan.Create;if len(existing)>0{if string(existing)==string(mb){mk=plan.Skip}else{mk=plan.Update}};ops=append(ops,plan.Operation{Kind:mk,Path:mp,BeforeDigest:fsx.BytesDigest(existing),AfterDigest:fsx.BytesDigest(mb),Payload:mb});return plan.New("init",root,now,ops),m,nil
}
