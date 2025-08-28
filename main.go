package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	flagName    = flag.String("name", "", "Project directory name (phx.new target)")
	flagApp     = flag.String("app", "", "Phoenix app name (e.g. my_app)")
	flagTimeout = flag.Duration("timeout", 30*time.Minute, "Overall timeout")

	flagListTpls = flag.Bool("list-templates", false, "List embedded templates and exit")
)

func main() {
	flag.Parse()

	tpl := NewEmbeddedTemplateEngine()
	if *flagListTpls {
		names, err := tpl.List()
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
		fmt.Fprintln(os.Stderr, "Usage: ./phx-bootstrap -name <dir> -app <app>")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
	defer cancel()

	td := &TemplateData{
		AppName:      *flagApp,
		AppModule:    toModule(*flagApp),
		AppModuleWeb: toModule(*flagApp) + "Web",
		ProjectName:  *flagName,
		// ProjectAbsDir is set later by the manifest via "resolve_project_dir"
	}

	flow := NewManifestFlow(ManifestFlowConfig{
		Templates: tpl,
		Exec:      SystemRunner{},
	})

	if err := flow.Run(ctx, td); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Done!")
}
