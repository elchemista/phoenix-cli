package manifest

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

	"github.com/elchemista/phoenix-cli/internal/assets"
	"github.com/elchemista/phoenix-cli/internal/templates"
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
	Type string `json:"type"` // "run" | "shell" | "overwrite" | "create" | "append" | "insert" | "resolve_project_dir"

	// For "run" (generic executable, e.g. "mix"):
	Name string   `json:"name,omitempty"`
	Args []string `json:"args,omitempty"`

	// For "shell":
	Script string `json:"script,omitempty"`

	// Optional working dir for run/shell (templated):
	Dir string `json:"dir,omitempty"`

	// For template file ops:
	Template  string `json:"template,omitempty"`
	Dest      string `json:"dest,omitempty"`
	Module    string `json:"module,omitempty"`
	Overwrite bool   `json:"overwrite,omitempty"` // used by "create"
}

func ExecuteManifest(ctx context.Context, exec Runner, t templates.TemplateEngine, td *templates.TemplateData) error {
	const manifestPath = "templates/manifest.json"

	// Read from the embedded assets FS explicitly.
	b, err := fs.ReadFile(assets.Templates, manifestPath)
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

func runAction(ctx context.Context, exec Runner, t templates.TemplateEngine, td *templates.TemplateData, a Action) error {
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

	case "overwrite":
		dest, err := applyMiniTemplate(a.Dest, *td)
		if err != nil {
			return err
		}
		full, err := resolveProjectPath(td, dest)
		if err != nil {
			return err
		}
		fmt.Printf("→ overwrite %s from template %s\n", full, a.Template)
		return t.CreateFileFromTemplate(full, a.Template, *td, true)

	case "create":
		dest, err := applyMiniTemplate(a.Dest, *td)
		if err != nil {
			return err
		}
		full, err := resolveProjectPath(td, dest)
		if err != nil {
			return err
		}
		ow := a.Overwrite // default false unless set in manifest
		fmt.Printf("→ create %s (overwrite=%v) from template %s\n", full, ow, a.Template)
		return t.CreateFileFromTemplate(full, a.Template, *td, ow)

	case "append":
		dest, err := applyMiniTemplate(a.Dest, *td)
		if err != nil {
			return err
		}
		full, err := resolveProjectPath(td, dest)
		if err != nil {
			return err
		}
		fmt.Printf("→ append %s from template %s\n", full, a.Template)
		return t.AppendTemplateToFile(full, a.Template, *td)

	case "insert":
		dest, err := applyMiniTemplate(a.Dest, *td)
		if err != nil {
			return err
		}
		mod, err := applyMiniTemplate(a.Module, *td)
		if err != nil {
			return err
		}
		full, err := resolveProjectPath(td, dest)
		if err != nil {
			return err
		}
		fmt.Printf("→ insert into %s (module %s) from template %s\n", full, mod, a.Template)
		return t.InsertTemplateBeforeModuleEnd(full, mod, a.Template, *td)

	default:
		return fmt.Errorf("unknown action type: %q", a.Type)
	}
}

func applyMiniTemplate(s string, td templates.TemplateData) (string, error) {
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
	return "sh", []string{"-lc"}
}

// ensure the project dir exists & is set
func ensureProjectDir(td *templates.TemplateData) error {
	if strings.TrimSpace(td.ProjectAbsDir) == "" {
		return fmt.Errorf("ProjectAbsDir is empty. Add a 'resolve_project_dir' action before file operations")
	}
	info, err := os.Stat(td.ProjectAbsDir)
	if err != nil {
		return fmt.Errorf("project dir not found: %s (%w)", td.ProjectAbsDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project dir is not a directory: %s", td.ProjectAbsDir)
	}
	return nil
}

// join <project>/ + dest (relative), prevent abs paths & traversal
func resolveProjectPath(td *templates.TemplateData, dest string) (string, error) {
	if err := ensureProjectDir(td); err != nil {
		return "", err
	}
	dest = filepath.FromSlash(strings.TrimSpace(dest))
	if dest == "" {
		return "", fmt.Errorf("empty dest")
	}
	if filepath.IsAbs(dest) {
		return "", fmt.Errorf("dest must be relative to project dir: %s", dest)
	}
	full := filepath.Clean(filepath.Join(td.ProjectAbsDir, dest))

	rel, err := filepath.Rel(td.ProjectAbsDir, full)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") || rel == "." {
		return "", fmt.Errorf("dest escapes project dir: %s -> %s", dest, full)
	}
	return full, nil
}
