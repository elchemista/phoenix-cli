package app

import (
	"github.com/elchemista/phoenix-cli/internal/assets"
	"github.com/elchemista/phoenix-cli/internal/manifest"
	"github.com/elchemista/phoenix-cli/internal/templates"
)

type App struct {
	Templates templates.TemplateEngine
	Exec      manifest.Runner
}

func New() *App {
	return &App{
		Templates: templates.NewEngine(assets.Templates, "templates"),
		Exec:      manifest.SystemRunner{},
	}
}
