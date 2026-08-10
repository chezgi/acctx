package skills

import (
	"acctx/internal/fsx"
	"acctx/internal/manifest"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Status struct {
	ID string `json:"id"`
	Version string `json:"version"`
	VendorPath string `json:"vendor_path"`
	ActivePath string `json:"active_path"`
	Digest string `json:"digest"`
	ActualDigest string `json:"actual_digest"`
	Override bool `json:"override"`
	Valid bool `json:"valid"`
	ProviderLinks map[string]string `json:"provider_links"`
}
type Service struct{ Root string }

func New(root string) Service { return Service{root} }
func (s Service) Status(id string) (Status, error) {
	m, e := manifest.Load(s.Root)
	if e != nil { return Status{}, e }
	var x *manifest.Skill
	for i := range m.Managed.Skills { if m.Managed.Skills[i].ID == id { x = &m.Managed.Skills[i]; break } }
	if x == nil { return Status{}, fmt.Errorf("skill not found") }
	ap := filepath.Join(s.Root, filepath.FromSlash(x.ActivePath))
	meta, me := ReadMetadata(ap)
	actual, e := treeDigest(ap)
	valid := me == nil && e == nil && meta.ID == x.ID && meta.Version == x.Version
	for prov, target := range x.ProviderLinks {
		lp := filepath.Join(s.Root, prov, id)
		got, e := os.Readlink(lp)
		if e != nil || got != target { valid = false }
	}
	return Status{x.ID, x.Version, x.VendorPath, x.ActivePath, x.Digest, actual, x.Override != nil, valid, x.ProviderLinks}, nil
}
func treeDigest(root string) (string, error) {
	m := map[string][]byte{}
	e := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil { return e }
		if d.IsDir() { return nil }
		r, _ := filepath.Rel(root, p)
		if filepath.Base(p) == ".acctx-override.yaml" { return nil }
		b, e := os.ReadFile(p); if e != nil { return e }
		m[filepath.ToSlash(r)] = b
		return nil
	})
	if e != nil { return "", e }
	ks := make([]string,0,len(m)); for k := range m { ks=append(ks,k) }; sort.Strings(ks)
	buf := []byte{}; for _,k := range ks { buf=append(buf,[]byte(k)...); buf=append(buf,0); buf=append(buf,m[k]...); buf=append(buf,0) }
	return fsx.BytesDigest(buf), nil
}
func (s Service) List() ([]Status,error) {
	m,e:=manifest.Load(s.Root); if e!=nil{return nil,e}
	out:=[]Status{}; for _,x:=range m.Managed.Skills { st,_:=s.Status(x.ID); out=append(out,st) }
	sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID}); return out,nil
}
