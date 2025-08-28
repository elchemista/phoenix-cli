package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

type Runner interface {
	Run(ctx context.Context, dir, name string, args ...string) error
}

type SystemRunner struct{}

func (SystemRunner) Run(ctx context.Context, dir, name string, args ...string) error {
	fmt.Printf("→ %s %s (in %s)\n", name, strings.Join(args, " "), dir)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = dir
	return cmd.Run()
}

type Manifest struct {
	Actions []Action `json:"actions"`
}

type Action struct {
	// Common:
	Type string `json:"type"` // "run" | "shell" | "overwrite" | "append" | "insert" | "resolve_project_dir"

	// For "run" (generic executable, e.g. "mix"):
	Name string   `json:"name,omitempty"` // e.g. "mix"
	Args []string `json:"args,omitempty"` // e.g. ["phx.new", "{{ .ProjectName }}", "--app", "{{ .AppName }}", "--no-install"]

	// For "shell":
	Script string `json:"script,omitempty"` // e.g. "cd lib/{{ .AppName }}_web && mkdir ciao"

	// Optional working dir for run/shell (templated):
	Dir string `json:"dir,omitempty"` // e.g. "{{ .ProjectAbsDir }}"

	// For template file ops:
	Template  string `json:"template,omitempty"`  // template file name under templates/
	Dest      string `json:"dest,omitempty"`      // destination path (templated)
	Module    string `json:"module,omitempty"`    // for "insert" (templated)
	Overwrite bool   `json:"overwrite,omitempty"` // for "overwrite" (defaults true in our flow)
}

func ExecuteManifest(ctx context.Context, exec Runner, t TemplateEngine, td *TemplateData) error {
	ete, ok := t.(*EmbeddedTemplateEngine)
	if !ok {
		return fmt.Errorf("unexpected template engine type: need EmbeddedTemplateEngine")
	}

	const manifestPath = "templates/manifest.json"
	b, err := fs.ReadFile(ete.FS, manifestPath)
	if err != nil {
		return fmt.Errorf("manifest missing: %s (%w)", manifestPath, err)
	}

	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("manifest parse error: %w", err)
	}

	for i, a := range m.Actions {
		if err := runAction(ctx, exec, t, td, a); err != nil {
			return fmt.Errorf("action %d (%s) failed: %w", i, a.Type, err)
		}
	}
	return nil
}

func runAction(ctx context.Context, exec Runner, t TemplateEngine, td *TemplateData, a Action) error {
	switch strings.ToLower(a.Type) {

	case "resolve_project_dir":
		abs, err := filepath.Abs(td.ProjectName)
		if err != nil {
			return err
		}
		td.ProjectAbsDir = abs
		return nil

	case "run":
		name, err := applyMiniTemplate(a.Name, *td)
		if err != nil {
			return err
		}
		dir, err := applyMiniTemplate(defaultIfEmpty(a.Dir, "."), *td)
		if err != nil {
			return err
		}
		args := make([]string, 0, len(a.Args))
		for _, raw := range a.Args {
			v, err := applyMiniTemplate(raw, *td)
			if err != nil {
				return err
			}
			args = append(args, v)
		}
		return exec.Run(ctx, dir, name, args...)

	case "shell":
		script, err := applyMiniTemplate(a.Script, *td)
		if err != nil {
			return err
		}
		dir, err := applyMiniTemplate(defaultIfEmpty(a.Dir, "."), *td)
		if err != nil {
			return err
		}
		sh, shArgs := systemShell()
		all := append(shArgs, script)
		return exec.Run(ctx, dir, sh, all...)

	case "overwrite", "create":
		dest, err := applyMiniTemplate(a.Dest, *td)
		if err != nil {
			return err
		}
		if err := t.CreateFileFromTemplate(filepath.Join(td.ProjectAbsDir, dest), a.Template, *td, true /*overwrite*/); err != nil {
			return err
		}
		return nil

	case "append":
		dest, err := applyMiniTemplate(a.Dest, *td)
		if err != nil {
			return err
		}
		if err := t.AppendTemplateToFile(filepath.Join(td.ProjectAbsDir, dest), a.Template, *td); err != nil {
			return err
		}
		return nil

	case "insert":
		dest, err := applyMiniTemplate(a.Dest, *td)
		if err != nil {
			return err
		}
		mod, err := applyMiniTemplate(a.Module, *td)
		if err != nil {
			return err
		}
		if err := t.InsertTemplateBeforeModuleEnd(filepath.Join(td.ProjectAbsDir, dest), mod, a.Template, *td); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown action type: %q", a.Type)
	}
}

func applyMiniTemplate(s string, td TemplateData) (string, error) {
	tpl, err := template.New("inline").Parse(s)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := tpl.Execute(&b, td); err != nil {
		return "", err
	}
	return b.String(), nil
}

func defaultIfEmpty(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func systemShell() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C"}
	}
	// POSIX
	return "sh", []string{"-lc"}
}
