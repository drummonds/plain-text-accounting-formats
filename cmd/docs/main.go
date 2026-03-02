package main

import (
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"sort"

	beancount "github.com/drummonds/gts-beancount"
	"github.com/odvcencio/gotreesitter"
)

type fileData struct {
	Name        string
	Source      string
	Highlighted template.HTML
	SExpr       string
	OK          bool
}

// captureToClass maps tree-sitter capture names to CSS classes.
var captureToClass = map[string]string{
	"keyword":          "hl-keyword",
	"string":           "hl-string",
	"number":           "hl-number",
	"constant":         "hl-constant",
	"constant.builtin": "hl-constant",
	"type":             "hl-type",
	"tag":              "hl-tag",
	"attribute":        "hl-attribute",
	"comment":          "hl-comment",
	"operator":         "hl-operator",
}

func renderHTML(src []byte, ranges []gotreesitter.HighlightRange) string {
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].StartByte < ranges[j].StartByte
	})

	var out []byte
	pos := uint32(0)
	for _, r := range ranges {
		if r.StartByte > pos {
			out = append(out, []byte(html.EscapeString(string(src[pos:r.StartByte])))...)
		}
		cls := captureToClass[r.Capture]
		text := html.EscapeString(string(src[r.StartByte:r.EndByte]))
		if cls != "" {
			out = fmt.Appendf(out, `<span class="%s">%s</span>`, cls, text)
		} else {
			out = append(out, []byte(text)...)
		}
		pos = r.EndByte
	}
	if pos < uint32(len(src)) {
		out = append(out, []byte(html.EscapeString(string(src[pos:])))...)
	}
	return string(out)
}

const tmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>gts-beancount — Parser Demo</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
<style>
  pre { background: #1e1e2e; color: #cdd6f4; padding: 1rem; border-radius: 6px; overflow-x: auto; font-size: 0.85rem; }
  .hl-keyword  { color: #cba6f7; font-weight: bold; }
  .hl-string   { color: #a6e3a1; }
  .hl-number   { color: #f9e2af; }
  .hl-constant { color: #89dceb; }
  .hl-type     { color: #89b4fa; font-weight: bold; }
  .hl-comment  { color: #6c7086; font-style: italic; }
  .hl-operator { color: #f38ba8; }
  .hl-tag      { color: #74c7ec; }
  .hl-attribute { color: #74c7ec; }
  details summary { cursor: pointer; font-weight: 600; margin-top: 0.5rem; }
</style>
</head>
<body>
<section class="section">
<div class="container">
  <h1 class="title">gts-beancount — Parser Demo</h1>
  {{range .}}
  <div class="box mb-5">
    <h2 class="subtitle">
      {{.Name}}
      {{if .OK}}<span class="tag is-success">OK</span>{{else}}<span class="tag is-danger">ERRORS</span>{{end}}
    </h2>
    <div class="columns">
      <div class="column">
        <h3 class="heading">Source</h3>
        <pre>{{.Source}}</pre>
      </div>
      <div class="column">
        <h3 class="heading">Highlighted</h3>
        <pre>{{.Highlighted}}</pre>
      </div>
    </div>
    <details>
      <summary>AST</summary>
      <pre>{{.SExpr}}</pre>
    </details>
  </div>
  {{end}}
</div>
</section>
</body>
</html>
`

func main() {
	files, err := filepath.Glob("testdata/*.beancount")
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(files)

	var data []fileData
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
			os.Exit(1)
		}

		stripped := beancount.StripBlankLines(src)

		tree, err := beancount.Parse(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err)
			os.Exit(1)
		}

		ranges, err := beancount.Highlight(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "highlight %s: %v\n", path, err)
			os.Exit(1)
		}

		data = append(data, fileData{
			Name:        filepath.Base(path),
			Source:      html.EscapeString(string(stripped)),
			Highlighted: template.HTML(renderHTML(stripped, ranges)),
			SExpr:       beancount.SExpression(tree),
			OK:          !beancount.HasErrors(tree),
		})
	}

	t, err := template.New("page").Parse(tmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "template: %v\n", err)
		os.Exit(1)
	}

	out, err := os.Create("docs/index.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	if err := t.Execute(out, data); err != nil {
		fmt.Fprintf(os.Stderr, "execute: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("wrote docs/index.html")
}
