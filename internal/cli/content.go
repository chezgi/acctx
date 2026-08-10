package cli

import (
	"acctx/internal/output"
	"fmt"
)

func runContent(a *app, args []string) error {
	if len(args) == 0 { return fmt.Errorf("content subcommand required") }
	switch args[0] {
	case "version":
		return output.Write(a.streams.Out,a.json,output.Result{Command:"content version",Status:"ok",Data:map[string]any{"content_version":a.bundle.Catalog.ContentVersion}})
	case "list":
		if a.json { return output.Write(a.streams.Out,true,output.Result{Command:"content list",Status:"ok",Data:a.bundle.Catalog}) }
		return output.Write(a.streams.Out,false,output.Result{Command:"content list",Status:"ok",Data:fmt.Sprintf("سال‌های پشتیبانی‌شده: %v | preset: %s | skills: %d",a.bundle.Catalog.SupportedYears,a.bundle.Catalog.DefaultPreset,len(a.bundle.Catalog.Skills))})
	default:
		return fmt.Errorf("unknown content subcommand")
	}
}
