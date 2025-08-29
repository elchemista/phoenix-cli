package templates

import (
	"os"
	"path/filepath"
	"strings"
)

// Helpers used by the template engine implementation

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

// ToModule converts "my_app-name" → "MyAppName"
func ToModule(app string) string {
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
