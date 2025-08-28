package main

import (
	"context"
)

type Step interface {
	Run(ctx context.Context, td *TemplateData) error
}

type Flow struct {
	steps []Step
}

func (f Flow) Run(ctx context.Context, td *TemplateData) error {
	for _, s := range f.steps {
		if err := s.Run(ctx, td); err != nil {
			return err
		}
	}
	return nil
}

type ManifestFlowConfig struct {
	Templates TemplateEngine
	Exec      Runner
}

func NewManifestFlow(cfg ManifestFlowConfig) Flow {
	return Flow{steps: []Step{
		StepExecuteManifest{Templates: cfg.Templates, Exec: cfg.Exec},
	}}
}

type StepExecuteManifest struct {
	Templates TemplateEngine
	Exec      Runner
}

func (s StepExecuteManifest) Run(ctx context.Context, td *TemplateData) error {
	return ExecuteManifest(ctx, s.Exec, s.Templates, td)
}
