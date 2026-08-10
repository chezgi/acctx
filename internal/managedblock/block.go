package managedblock

import (
	"bytes"
	"errors"
	"fmt"
)

var ErrConflict = errors.New("managed block conflict")

type Style struct{ Begin, End string }

var Markdown = Style{"<!-- acctx:begin -->", "<!-- acctx:end -->"}
var Gitignore = Style{"# <!-- acctx:begin -->", "# <!-- acctx:end -->"}

type Result struct {
	Content []byte
	Changed bool
}

func Merge(existing, body []byte, s Style) (Result, error) {
	bc := bytes.Count(existing, []byte(s.Begin))
	ec := bytes.Count(existing, []byte(s.End))
	if bc != ec || bc > 1 {
		return Result{}, fmt.Errorf("%w: invalid marker count", ErrConflict)
	}
	nl := "\n"
	if bytes.Contains(existing, []byte("\r\n")) {
		nl = "\r\n"
	}
	body = bytes.TrimRight(body, "\r\n")
	block := []byte(s.Begin + nl + string(body) + nl + s.End)
	var out []byte
	if bc == 1 {
		a := bytes.Index(existing, []byte(s.Begin))
		b := bytes.Index(existing, []byte(s.End))
		if b < a {
			return Result{}, fmt.Errorf("%w: marker order", ErrConflict)
		}
		out = append(out, existing[:a]...)
		out = append(out, block...)
		out = append(out, existing[b+len(s.End):]...)
	} else {
		out = append(out, existing...)
		if len(out) > 0 && !bytes.HasSuffix(out, []byte(nl)) {
			out = append(out, []byte(nl)...)
		}
		if len(out) > 0 && !bytes.HasSuffix(out, []byte(nl+nl)) {
			out = append(out, []byte(nl)...)
		}
		out = append(out, block...)
		out = append(out, []byte(nl)...)
	}
	return Result{out, !bytes.Equal(out, existing)}, nil
}
func Body(existing []byte, s Style) ([]byte, error) {
	bc := bytes.Count(existing, []byte(s.Begin))
	ec := bytes.Count(existing, []byte(s.End))
	if bc != 1 || ec != 1 {
		return nil, ErrConflict
	}
	a := bytes.Index(existing, []byte(s.Begin)) + len(s.Begin)
	b := bytes.Index(existing, []byte(s.End))
	if b < a {
		return nil, ErrConflict
	}
	return bytes.Trim(existing[a:b], "\r\n"), nil
}
