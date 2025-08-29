package flow

import (
	"context"

	"github.com/elchemista/phoenix-cli/internal/manifest"
	"github.com/elchemista/phoenix-cli/internal/templates"
)

type Step interface {
	Run(ctx context.Context, td *templates.TemplateData) error
}

type Flow struct {
	steps []Step
}

func (f Flow) Run(ctx context.Context, td *templates.TemplateData) error {
	for _, s := range f.steps {
		if err := s.Run(ctx, td); err != nil {
			return err
		}
	}
	return nil
}

type ManifestFlowConfig struct {
	Templates templates.TemplateEngine
	Exec      manifest.Runner
}

func NewManifestFlow(cfg ManifestFlowConfig) Flow {
	return Flow{steps: []Step{
		StepExecuteManifest{Templates: cfg.Templates, Exec: cfg.Exec},
	}}
}

type StepExecuteManifest struct {
	Templates templates.TemplateEngine
	Exec      manifest.Runner
}

func (s StepExecuteManifest) Run(ctx context.Context, td *templates.TemplateData) error {
	return manifest.ExecuteManifest(ctx, s.Exec, s.Templates, td)
}
