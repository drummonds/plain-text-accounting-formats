package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/drummonds/gotreesitter"
	"github.com/drummonds/gotreesitter/grammars"
	beancount "github.com/drummonds/plain-text-accounting-formats"
)

type formatPage struct {
	Name        string
	Slug        string
	Heading     string        // subtitle for format page
	Description template.HTML // HTML description for format page
	CardDesc    string        // short description for index card
	GrammarJS   template.HTML
	GrammarJSON template.HTML
	ABNF        template.HTML
	RoundTrip   template.HTML
	Demos       []demoData
}

type demoData struct {
	Name               string
	BeancountHighlight template.HTML
	BeancountOK        bool
	GolucaHighlight    template.HTML
	GolucaOK           bool
	SExpr              string
	SVGContent         template.HTML
}

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
<title>Parser Demo — Plain Text Accounting Formats</title>
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
  .svg-container svg { max-width: 100%; height: auto; }
`

var indexTmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Plain Text Accounting Formats</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
<style>` + sharedCSS + `</style>
</head>
<body>
<section class="section">
<div class="container">
  <h1 class="title">Parser Demo — Plain Text Accounting Formats</h1>
  {{range .}}
  <div class="box mb-5">
    <h2 class="subtitle">{{.Name}}</h2>
    <div class="columns">
      <div class="column">
        <h3 class="heading">Beancount {{if .BeancountOK}}<span class="tag is-success">OK</span>{{else}}<span class="tag is-danger">ERRORS</span>{{end}}</h3>
        <pre>{{.BeancountHighlight}}</pre>
      </div>
      <div class="column">
        <h3 class="heading">Goluca {{if .GolucaOK}}<span class="tag is-success">OK</span>{{else}}<span class="tag is-danger">ERRORS</span>{{end}}</h3>
        <pre>{{.GolucaHighlight}}</pre>
      </div>
    </div>
    {{end}}
  </div>
</div>
</section>
</body>
</html>
`

var formatTmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>{{.Name}} — Plain Text Accounting Formats</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
<style>` + sharedCSS + `</style>
</head>
<body>
<section class="section">
<div class="container">
  <nav class="breadcrumb" aria-label="breadcrumbs">
    <ul>
      <li><a href="index.html">Formats</a></li>
      <li class="is-active"><a href="#">{{.Name}}</a></li>
    </ul>
  </nav>

  <h1 class="title">{{.Name}}</h1>
  <p class="subtitle">{{.Heading}}</p>

  <div class="content mb-5">
    <p>{{.Description}}</p>
  </div>

  <div class="box">
    <h2 class="subtitle">ABNF Grammar</h2>
    <pre>{{.ABNF}}</pre>
    <details>
      <summary>ABNF → JSON round-trip</summary>
      <pre>{{.RoundTrip}}</pre>
    </details>
  </div>

  <div class="box">
    <h2 class="subtitle">Source Files</h2>
    <details>
      <summary>grammar.js</summary>
      <pre>{{.GrammarJS}}</pre>
    </details>
    <details>
      <summary>grammar.json</summary>
      <pre>{{.GrammarJSON}}</pre>
    </details>
  </div>

  {{if .Demos}}
  <h2 class="subtitle">Parser Demos</h2>
  {{range .Demos}}
  <div class="box mb-5">
    <h3 class="subtitle is-5">{{.Name}}</h3>
    <div class="columns">
      <div class="column">
        <p class="heading">Beancount {{if .BeancountOK}}<span class="tag is-success is-light">OK</span>{{else}}<span class="tag is-danger is-light">ERRORS</span>{{end}}</p>
        <pre>{{.BeancountHighlight}}</pre>
      </div>
      <div class="column">
        <p class="heading">Goluca {{if .GolucaOK}}<span class="tag is-success is-light">OK</span>{{else}}<span class="tag is-danger is-light">ERRORS</span>{{end}}</p>
        <pre>{{.GolucaHighlight}}</pre>
      </div>
    </div>
    {{if .SVGContent}}
    <div class="svg-container mt-3">
      <p class="heading">Flow Diagram</p>
      {{.SVGContent}}
    </div>
    {{end}}
    <details>
      <summary>AST (beancount)</summary>
      <pre>{{.SExpr}}</pre>
    </details>
  </div>
  {{end}}
  {{end}}

		golucaTree, err := beancount.GolucaParse(golucaSrc)
		golucaOK := err == nil && golucaTree != nil && !golucaTree.RootNode().HasError()

		var golucaHL template.HTML
		golucaRanges, err := beancount.GolucaHighlight(golucaSrc)
		if err == nil {
			golucaHL = template.HTML(renderHTML(golucaSrc, golucaRanges))
		} else {
			golucaHL = template.HTML(html.EscapeString(string(golucaSrc)))
		}

		data = append(data, fileData{
			Name:               filepath.Base(path),
			BeancountHighlight: template.HTML(renderHTML(stripped, bcRanges)),
			BeancountOK:        !beancount.HasErrors(tree),
			GolucaSource:       string(golucaSrc),
			GolucaHighlight:    golucaHL,
			GolucaOK:           golucaOK,
			SExpr:              beancount.SExpression(tree),
		})
	}

	t, err := template.New("page").Parse(tmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "template: %v\n", err)
		os.Exit(1)
	}

	out, err := os.Create("docs/demo.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	if err := t.Execute(out, data); err != nil {
		fmt.Fprintf(os.Stderr, "execute: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("wrote docs/demo.html")
}
