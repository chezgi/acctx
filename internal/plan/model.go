package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"time"
)

type Kind string

const (
	Create   Kind = "CREATE"
	Update   Kind = "UPDATE"
	Link     Kind = "LINK"
	Delete   Kind = "DELETE"
	Skip     Kind = "SKIP"
	Conflict Kind = "CONFLICT"
)

type Operation struct {
	Kind         Kind   `json:"kind" yaml:"kind"`
	Path         string `json:"path" yaml:"path"`
	Target       string `json:"target,omitempty" yaml:"target,omitempty"`
	BeforeDigest string `json:"before_digest,omitempty" yaml:"before_digest,omitempty"`
	AfterDigest  string `json:"after_digest,omitempty" yaml:"after_digest,omitempty"`
	Message      string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload      []byte `json:"-" yaml:"-"`
}
type Summary struct{ Creates, Updates, Links, Deletes, Skips, Conflicts int }
type Plan struct {
	ID         string      `json:"id" yaml:"id"`
	Command    string      `json:"command" yaml:"command"`
	Root       string      `json:"root" yaml:"root"`
	CreatedAt  time.Time   `json:"created_at" yaml:"created_at"`
	Operations []Operation `json:"operations" yaml:"operations"`
	Summary    Summary     `json:"summary" yaml:"summary"`
}

func New(cmd, root string, now time.Time, ops []Operation) Plan {
	sort.SliceStable(ops, func(i, j int) bool {
		if ops[i].Path == ops[j].Path {
			return ops[i].Kind < ops[j].Kind
		}
		return ops[i].Path < ops[j].Path
	})
	h := sha256.Sum256([]byte(cmd + root + now.UTC().Format(time.RFC3339Nano)))
	p := Plan{ID: "op-" + hex.EncodeToString(h[:6]), Command: cmd, Root: root, CreatedAt: now.UTC(), Operations: ops}
	for _, o := range ops {
		switch o.Kind {
		case Create:
			p.Summary.Creates++
		case Update:
			p.Summary.Updates++
		case Link:
			p.Summary.Links++
		case Delete:
			p.Summary.Deletes++
		case Skip:
			p.Summary.Skips++
		case Conflict:
			p.Summary.Conflicts++
		}
	}
	return p
}
func (p Plan) HasConflicts() bool { return p.Summary.Conflicts > 0 }
func (p Plan) HasChanges() bool {
	return p.Summary.Creates+p.Summary.Updates+p.Summary.Links+p.Summary.Deletes > 0
}
