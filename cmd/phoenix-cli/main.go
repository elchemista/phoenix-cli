package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/elchemista/phoenix-cli/internal/app"
	"github.com/elchemista/phoenix-cli/internal/flow"
	"github.com/elchemista/phoenix-cli/internal/templates"
)

var (
	flagName    = flag.String("name", "", "Project directory name (phx.new target)")
	flagApp     = flag.String("app", "", "Phoenix app name (e.g. my_app)")
	flagTimeout = flag.Duration("timeout", 30*time.Minute, "Overall timeout")

	flagListTpls = flag.Bool("list-templates", false, "List embedded templates and exit")
)

func main() {
	flag.Parse()

	a := app.New() // wires: template engine + system runner

	if *flagListTpls {
		names, err := a.Templates.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, "template list error:", err)
			os.Exit(1)
		}
		for _, n := range names {
			fmt.Println(n)
		}
		return
	}

	if *flagName == "" || *flagApp == "" {
		fmt.Fprintln(os.Stderr, "Usage: phoenix-cli -name <dir> -app <app>")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
	defer cancel()

	td := &templates.TemplateData{
		AppName:      *flagApp,
		AppModule:    templates.ToModule(*flagApp),
		AppModuleWeb: templates.ToModule(*flagApp) + "Web",
		ProjectName:  *flagName,
		// ProjectAbsDir is set later by the manifest via "resolve_project_dir"
	}

	f := flow.NewManifestFlow(flow.ManifestFlowConfig{
		Templates: a.Templates,
		Exec:      a.Exec,
	})

	if err := f.Run(ctx, td); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Done!")
}
