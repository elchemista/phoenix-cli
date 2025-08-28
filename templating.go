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

func (e *EmbeddedTemplateEngine) Render(name string, data any) (string, error) {
	loc := path.Join(e.Base, name)
	t := template.New(name)
	if e.Funcs != nil {
		t = t.Funcs(e.Funcs)
	}
	parsed, err := t.ParseFS(e.FS, loc)
	if err != nil {
		return "", fmt.Errorf("template not embedded: %s (%w)", loc, err)
	}
	var buf bytes.Buffer
	if err := parsed.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
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
	return os.WriteFile(dstPath, []byte(out), 0o644)
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
	defer f.Close()
	_, err = f.WriteString(out)
	return err
}

func (e *EmbeddedTemplateEngine) InsertTemplateBeforeModuleEnd(dstPath, fullModule, tplName string, data any) error {
	out, err := e.Render(tplName, data)
	if err != nil {
		return err
	}
	return insertBeforeModuleEnd(dstPath, fullModule, "\n"+out+"\n")
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
		fmt.Printf("! warning: module header not found (%s). Appending.\n", fullModule)
		return appendRaw(path, "\n"+content)
	}

	trimmed := strings.TrimRight(s, " \t\r\n")
	last := strings.LastIndex(trimmed, "\nend")
	if last == -1 {
		fmt.Println("! warning: final 'end' not found. Appending.")
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
