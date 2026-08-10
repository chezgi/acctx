package output

import (
	"acctx/internal/diagnostic"
	"encoding/json"
	"fmt"
	"io"
)

type Result struct {
	Command string `json:"command"`
	Status string `json:"status"`
	Data any `json:"data,omitempty"`
	Diagnostics []diagnostic.Item `json:"diagnostics,omitempty"`
}

func Write(w io.Writer, asJSON bool, r Result) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	}
	if r.Status == "error" {
		for _, d := range r.Diagnostics {
			if _, err := fmt.Fprintf(w, "%s: %s\n", d.Code, d.Message); err != nil {
				return err
			}
		}
		return nil
	}
	if r.Data != nil {
		_, err := fmt.Fprintf(w, "%v\n", r.Data)
		return err
	}
	for _, d := range r.Diagnostics {
		if _, err := fmt.Fprintf(w, "%s\n", d.Message); err != nil {
			return err
		}
	}
	return nil
}
