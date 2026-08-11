package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/output"
	"acctx/internal/plan"
	"acctx/internal/project"
	"acctx/internal/workspace"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type bootstrap struct {
	SchemaVersion int    `json:"schema_version"`
	ProjectID     string `json:"project_id"`
	Preset        string `json:"preset"`
}

func resolveGitRoot(a *app) (string, error) {
	start := a.root
	if start == "" {
		start, _ = os.Getwd()
	}
	root, err := project.GitRoot(start)
	if err != nil {
		return "", &ExitError{
			Code: ExitConfig,
			Diagnostic: diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_GIT_REPOSITORY_NOT_FOUND",
				Message:  "Git repository پیدا نشد",
			},
		}
	}
	return root, nil
}

func resolveProjectRoot(a *app) (string, error) {
	start := a.root
	if start == "" {
		start, _ = os.Getwd()
	}
	root, err := project.Root(start)
	if err != nil {
		return "", &ExitError{
			Code: ExitConfig,
			Diagnostic: diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_PROJECT_NOT_FOUND",
				Message:  "پروژه acctx پیدا نشد",
			},
		}
	}
	return root, nil
}

func renderPlan(a *app, operationPlan plan.Plan) error {
	if a.json {
		return output.Write(a.streams.Out, true, output.Result{Command: operationPlan.Command, Status: "ok", Data: operationPlan})
	}
	for _, operation := range operationPlan.Operations {
		fmt.Fprintf(a.streams.Out, "%-8s %s", operation.Kind, operation.Path)
		if operation.Target != "" {
			fmt.Fprintf(a.streams.Out, " -> %s", operation.Target)
		}
		if operation.Message != "" {
			fmt.Fprintf(a.streams.Out, " (%s)", operation.Message)
		}
		fmt.Fprintln(a.streams.Out)
	}
	return nil
}

func confirm(a *app, yes bool) bool {
	if yes {
		return true
	}
	if !a.streams.Interactive {
		return false
	}
	fmt.Fprint(a.streams.Out, "اعمال شود؟ [y/N] ")
	value, _ := bufio.NewReader(a.streams.In).ReadString('\n')
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "y" || value == "yes"
}

func parseFlags(args []string) (map[string]string, map[string]bool, []string, error) {
	values := map[string]string{}
	booleans := map[string]bool{}
	rest := []string{}
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch argument {
		case "--plan", "--yes", "--non-interactive", "--allow-dirty", "--force", "--historical", "--archive", "--include-company", "--include-year-inputs":
			booleans[argument] = true
		default:
			if strings.HasPrefix(argument, "--") {
				if index+1 >= len(args) {
					return nil, nil, nil, fmt.Errorf("missing value for %s", argument)
				}
				values[argument] = args[index+1]
				index++
			} else {
				rest = append(rest, argument)
			}
		}
	}
	return values, booleans, rest, nil
}

func runInit(a *app, args []string) error {
	if err := a.linuxMutating(); err != nil {
		return err
	}
	values, flags, _, err := parseFlags(args)
	if err != nil {
		return err
	}
	root, err := resolveGitRoot(a)
	if err != nil {
		return err
	}
	projectID := values["--project-id"]
	preset := values["--preset"]
	if config := values["--config"]; config != "" {
		data, err := os.ReadFile(config)
		if err != nil {
			return err
		}
		var configured bootstrap
		if err := json.Unmarshal(data, &configured); err != nil {
			return fmt.Errorf("bootstrap config must be JSON-compatible YAML: %w", err)
		}
		if configured.SchemaVersion != 0 && configured.SchemaVersion != 1 {
			return &ExitError{Code: ExitConfig, Diagnostic: diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CONFIG_INVALID", Message: "config schema معتبر نیست"}}
		}
		if projectID == "" {
			projectID = configured.ProjectID
		}
		if preset == "" {
			preset = configured.Preset
		}
	}
	if projectID == "" {
		projectID = filepath.Base(root)
	}
	operationPlan, _, err := workspace.BuildProjectInitPlan(root, workspace.InitOptions{ProjectID: projectID, Preset: preset}, a.bundle, a.build, time.Now().UTC())
	if err != nil {
		return err
	}
	return applyCommandPlan(a, operationPlan, flags)
}
