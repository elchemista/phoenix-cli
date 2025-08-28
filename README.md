# phx-bootstrap

This tool automates the bootstrapping of new Phoenix projects by running mix phx.new, updating dependencies, and applying a set of pre-defined customizations.

With Phoenix 1.8, the framework has become intentionally minimal. While the core architecture is excellent, some essentials (like dialogs, common UI components, or even simple daisyUI integration) are no longer included.

I built this CLI because I don’t want to reinvent or copy-paste the same solutions into every project — things like modals, daisyUI components, or opinionated defaults should be one-command away.

### Why this matters

Phoenix 1.8 is leaner by design. That’s good for flexibility, but in practice, it often means every new project requires repetitive setup:

 * Re-adding dialogs/modals.

 * Wiring up DaisyUI.

 * Extending core_components.ex.

 * Adjusting JS pipeline.

 * Adding missing conveniences in auth.

This CLI automates all of that. It lets you start with a ready-to-use Phoenix + DaisyUI + your own conventions — instead of copy-pasting from your last project.

## Quick start

1. Put your steps in `templates/manifest.json` (see examples below).
2. Put any content templates in `templates/*.tpl`.
3. Build and run:

   ```bash
   go build -o phx-bootstrap .
   ./phx-bootstrap -name consulente -app consulente
   ```
4. (Optional) See what templates are embedded:

   ```bash
   ./phx-bootstrap -list-templates
   ```

---

## Template data you can interpolate

Anywhere the manifest supports templating (`name`, `args[]`, `dir`, `dest`, `module`, `script`) you can use Go `text/template` syntax with the following fields:

* `{{ .AppName }}` — e.g., `my_app`
* `{{ .AppModule }}` — e.g., `MyApp`
* `{{ .AppModuleWeb }}` — e.g., `MyAppWeb`
* `{{ .ProjectName }}` — directory to create with `mix phx.new`
* `{{ .ProjectAbsDir }}` — absolute path of the generated project (populated after you run `resolve_project_dir`)

> Tip: Use forward slashes in paths inside the manifest (they work cross-platform). Quotes are JSON quotes `"..."`, not shell quotes.

---

## Action catalog (what you can do)

Actions run **in order**, top to bottom.

### 1) `run` — run an executable (e.g., `mix`)

```json
{
  "type": "run",
  "name": "mix",
  "args": ["phx.new", "{{ .ProjectName }}", "--app", "{{ .AppName }}", "--no-install"],
  "dir": "."
}
```

* `name`: executable name
* `args`: argument list (templated)
* `dir`: working directory (templated; default `"."`)

### 2) `resolve_project_dir` — compute absolute dir

```json
{ "type": "resolve_project_dir" }
```

* Must come **after** `phx.new` and **before** any action that needs `{{ .ProjectAbsDir }}`.

### 3) `shell` — run a shell snippet (multi-command)

```json
{
  "type": "shell",
  "script": "cd lib/{{ .AppName }}_web && mkdir -p components/shared",
  "dir": "{{ .ProjectAbsDir }}"
}
```

* Runs via `sh -lc` on macOS/Linux, `cmd /C` on Windows.
* Prefer setting `dir` instead of `cd`, but both are fine.

### 4) `overwrite` / `create` — write a file from a template

```json
{
  "type": "overwrite",
  "template": "assets_js_app.js.tpl",
  "dest": "assets/js/app.js",
  "overwrite": true
}
```

* `template`: file under `templates/` (embedded at build time)
* `dest`: path **inside the project** (templated), joined to `{{ .ProjectAbsDir }}`
* Always overwrites when `overwrite: true`.

### 5) `append` — append a template to a file

```json
{
  "type": "append",
  "template": "readme_append.md.tpl",
  "dest": "README.md"
}
```

### 6) `insert` — insert a template **before the final `end`** of a module

```json
{
  "type": "insert",
  "template": "core_components_append.ex.tpl",
  "dest": "lib/{{ .AppName }}_web/components/core_components.ex",
  "module": "{{ .AppModuleWeb }}.CoreComponents"
}
```

* Looks for `defmodule <module> do` and inserts the rendered template **before** the file’s final `end`.
* If the module isn’t found, falls back to appending at end of file (with a warning).

---

## Full example `templates/manifest.json`

This reproduces a typical flow: create project, resolve path, update deps, customize code, rewrite app.js, and run a shell step.

```json
{
  "actions": [
    {
      "type": "run",
      "name": "mix",
      "args": ["phx.new", "{{ .ProjectName }}", "--app", "{{ .AppName }}", "--no-install"],
      "dir": "."
    },
    { "type": "resolve_project_dir" },
    {
      "type": "run",
      "name": "mix",
      "args": ["deps.update", "--all"],
      "dir": "{{ .ProjectAbsDir }}"
    },
    {
      "type": "insert",
      "template": "core_components_append.ex.tpl",
      "dest": "lib/{{ .AppName }}_web/components/core_components.ex",
      "module": "{{ .AppModuleWeb }}.CoreComponents"
    },
    {
      "type": "overwrite",
      "template": "assets_js_app.js.tpl",
      "dest": "assets/js/app.js",
      "overwrite": true
    },
    {
      "type": "append",
      "template": "readme_append.md.tpl",
      "dest": "README.md"
    },
    {
      "type": "shell",
      "script": "cd lib/{{ .AppName }}_web && mkdir -p ciao",
      "dir": "{{ .ProjectAbsDir }}"
    }
  ]
}
```

---

## Template files (`templates/*.tpl`)

All `.tpl` files in `templates/` are embedded into the binary at build time (via `go:embed templates/**`). They render with full `text/template` support and the same `TemplateData` (`.AppName`, `.AppModule`, etc.).

**Example: `templates/core_components_append.ex.tpl`**

```elixir
# -- Injected by CLI --
@doc "Demo component injected by CLI"
def demo_badge(assigns) do
  ~H"""
  <span class="badge badge-info">Hello from {{ .AppModule }}</span>
  """
end
```

**Example: `templates/assets_js_app.js.tpl`**

```js
import "phoenix_html"
import topbar from "topbar"
import {Socket} from "phoenix"
import {LiveSocket} from "phoenix_live_view"

const csrfToken = document.querySelector("meta[name='csrf-token']").getAttribute("content")
const hooks = {}

topbar.config({barThickness: 3})
window.addEventListener("phx:page-loading-start", () => topbar.show())
window.addEventListener("phx:page-loading-stop", () => topbar.hide())

const liveSocket = new LiveSocket("/live", Socket, { params: { _csrf_token: csrfToken }, hooks })
liveSocket.connect()
window.liveSocket = liveSocket
```

## Troubleshooting

* **“template not embedded”**
  Make sure the file exists in `templates/` **at build time** and rebuild. Check with `-list-templates`.

* **“project dir not resolved”**
  You used `{{ .ProjectAbsDir }}` before running `resolve_project_dir`. Move that action earlier.

* **Windows path issues**
  Keep manifest paths with forward slashes (`lib/{{ .AppName }}_web/...`). The tool joins them safely.

---

## Recap

* Put **all steps** in `templates/manifest.json`.
* Use `{{ .AppName }}`, `{{ .AppModule }}`, `{{ .AppModuleWeb }}`, `{{ .ProjectName }}`, `{{ .ProjectAbsDir }}` in any action field that supports templating.
* Place content `.tpl` files in `templates/`; they’re compiled into the binary.
* Build a single executable and automate your Phoenix bootstrap start-to-finish.
