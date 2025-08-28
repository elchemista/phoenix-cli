package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

// Data available to templates
type TemplateData struct {
	AppName       string // "my_app"
	AppModule     string // "MyApp"
	AppModuleWeb  string // "MyAppWeb"
	ProjectName   string // created by phx.new
	ProjectAbsDir string // absolute path after generation
}

// Embed templates/**
//
//go:embed templates/**
var embeddedTemplates embed.FS

// TemplateEngine API
type TemplateEngine interface {
	Render(name string, data any) (string, error)
	CreateFileFromTemplate(dstPath, tplName string, data any, overwrite bool) error
	AppendTemplateToFile(dstPath, tplName string, data any) error
	InsertTemplateBeforeModuleEnd(dstPath, fullModule, tplName string, data any) error
	List() ([]string, error)
}

// Implementation using embedded FS
type EmbeddedTemplateEngine struct {
	FS    fs.FS
	Base  string
	Funcs template.FuncMap
}

func NewEmbeddedTemplateEngine() *EmbeddedTemplateEngine {
	return &EmbeddedTemplateEngine{
		FS:   embeddedTemplates,
		Base: "templates",
	}
}

// minimal logger
func logf(format string, args ...any) { fmt.Fprintf(os.Stderr, "[tpl] "+format+"\n", args...) }

func (e *EmbeddedTemplateEngine) Render(name string, data any) (string, error) {
	loc := path.Join(e.Base, name)

	b, err := fs.ReadFile(e.FS, loc)
	if err != nil {
		return "", fmt.Errorf("template not embedded: %s (%w)", loc, err)
	}

	// Try Go template first; fall back to raw if parsing/executing fails
	root := "tpl"
	t := template.New(root)
	if e.Funcs != nil {
		t = t.Funcs(e.Funcs)
	}
	if _, err := t.Parse(string(b)); err == nil {
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err == nil {
			return buf.String(), nil
		}
	}
	return string(b), nil
}

func (e *EmbeddedTemplateEngine) CreateFileFromTemplate(dstPath, tplName string, data any, overwrite bool) error {
	if !overwrite && fileExists(dstPath) {
		return fmt.Errorf("file exists: %s (set overwrite=true to replace)", dstPath)
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	out, err := e.Render(tplName, data)
	if err != nil {
		return err
	}

	f, err := os.Create(dstPath) // explicit create/truncate
	if err != nil {
		return err
	}
	if _, err := f.WriteString(out); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	abs, _ := filepath.Abs(dstPath)
	logf("write %dB -> %s", len(out), abs)
	return nil
}

func (e *EmbeddedTemplateEngine) AppendTemplateToFile(dstPath, tplName string, data any) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	out, err := e.Render(tplName, data)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(out); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	abs, _ := filepath.Abs(dstPath)
	logf("append %dB -> %s", len(out), abs)
	return nil
}

func (e *EmbeddedTemplateEngine) InsertTemplateBeforeModuleEnd(dstPath, fullModule, tplName string, data any) error {
	out, err := e.Render(tplName, data)
	if err != nil {
		return err
	}
	if err := insertBeforeModuleEnd(dstPath, fullModule, "\n"+out+"\n"); err != nil {
		return err
	}
	abs, _ := filepath.Abs(dstPath)
	logf("insert %dB into %s (%s)", len(out), abs, fullModule)
	return nil
}

func (e *EmbeddedTemplateEngine) List() ([]string, error) {
	var names []string
	err := fs.WalkDir(e.FS, e.Base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		names = append(names, strings.TrimPrefix(p, e.Base+"/"))
		return nil
	})
	return names, err
}

// Utilities -------------------------------------------------------------------

func insertBeforeModuleEnd(path string, fullModule string, content string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(b)

	header := "defmodule " + fullModule + " do"
	if !strings.Contains(s, header) {
		// header not found; append at end
		return appendRaw(path, "\n"+content)
	}

	trimmed := strings.TrimRight(s, " \t\r\n")
	last := strings.LastIndex(trimmed, "\nend")
	if last == -1 {
		// no final end; append
		return appendRaw(path, "\n"+content)
	}

	newContent := trimmed[:last] + "\n" + content + "end" + s[len(trimmed):]
	return os.WriteFile(path, []byte(newContent), 0o644)
}

func appendRaw(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func toModule(app string) string {
	app = strings.TrimSpace(app)
	app = strings.ReplaceAll(app, "-", "_")
	parts := strings.Split(app, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}
