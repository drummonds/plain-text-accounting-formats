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

	"codeberg.org/hum3/gotreesitter"
	"codeberg.org/hum3/gotreesitter/grammars"
	beancount "github.com/drummonds/plain-text-accounting-formats"
)

// --- Types ---

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

// --- Tree-sitter highlighting (beancount/goluca demos) ---

var captureToClass = map[string]string{
	"keyword":          "hl-keyword",
	"string":           "hl-string",
	"string.special":   "hl-string",
	"number":           "hl-number",
	"constant":         "hl-constant",
	"constant.builtin": "hl-constant",
	"type":             "hl-type",
	"type.builtin":     "hl-type",
	"tag":              "hl-tag",
	"attribute":        "hl-attribute",
	"comment":          "hl-comment",
	"operator":         "hl-operator",
	"variable":         "hl-tag",
	"property":         "hl-attribute",
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

// --- Regex-based highlighting (ABNF, JS fallback) ---

type hlRule struct {
	re    *regexp.Regexp
	class string
}

// highlightText applies rules in priority order (earlier rules win on overlap).
func highlightText(src string, rules []hlRule) template.HTML {
	taken := make([]bool, len(src))
	type hlMatch struct {
		start, end int
		class      string
	}
	var matches []hlMatch

	for _, rule := range rules {
		for _, loc := range rule.re.FindAllStringIndex(src, -1) {
			overlap := false
			for i := loc[0]; i < loc[1]; i++ {
				if taken[i] {
					overlap = true
					break
				}
			}
			if !overlap {
				matches = append(matches, hlMatch{loc[0], loc[1], rule.class})
				for i := loc[0]; i < loc[1]; i++ {
					taken[i] = true
				}
			}
		}
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].start < matches[j].start })

	var b strings.Builder
	pos := 0
	for _, m := range matches {
		if m.start > pos {
			b.WriteString(html.EscapeString(src[pos:m.start]))
		}
		b.WriteString(`<span class="` + m.class + `">`)
		b.WriteString(html.EscapeString(src[m.start:m.end]))
		b.WriteString(`</span>`)
		pos = m.end
	}
	if pos < len(src) {
		b.WriteString(html.EscapeString(src[pos:]))
	}
	return template.HTML(b.String())
}

// highlightWithTS uses a gotreesitter grammar for syntax highlighting.
// Falls back to regex rules if tree-sitter produces no results.
func highlightWithTS(ext string, src string, fallback []hlRule) template.HTML {
	entry := grammars.DetectLanguage(ext)
	if entry != nil && entry.HighlightQuery != "" {
		if hl, err := gotreesitter.NewHighlighter(entry.Language(), entry.HighlightQuery); err == nil {
			if ranges := hl.Highlight([]byte(src)); len(ranges) > 0 {
				return template.HTML(renderHTML([]byte(src), ranges))
			}
		}
	}
	if len(fallback) > 0 {
		return highlightText(src, fallback)
	}
	return template.HTML(html.EscapeString(src))
}

// Regex fallback rules for JavaScript (grammar.js files with complex regex
// literals can trip up the tree-sitter JS parser).
var jsRules = []hlRule{
	{regexp.MustCompile(`(?s)/\*.*?\*/`), "hl-comment"},
	{regexp.MustCompile(`//[^\n]*`), "hl-comment"},
	{regexp.MustCompile(`/(?:[^/\\\n]|\\.)+/[gimsuy]*`), "hl-string"},
	{regexp.MustCompile("`[^`]*`"), "hl-string"},
	{regexp.MustCompile(`"(?:[^"\\]|\\.)*"`), "hl-string"},
	{regexp.MustCompile(`'(?:[^'\\]|\\.)*'`), "hl-string"},
	{regexp.MustCompile(`\b(module|exports|grammar|rules|extras|externals|inline|conflicts|precedences|supertypes|word|choice|seq|repeat|repeat1|optional|token|field|prec|alias|const|let|var|function|return|if|else|true|false|null)\b`), "hl-keyword"},
	{regexp.MustCompile(`\$\.\w+`), "hl-type"},
	{regexp.MustCompile(`=>`), "hl-operator"},
	{regexp.MustCompile(`\b\d+(?:\.\d+)?\b`), "hl-number"},
}

// --- Format definitions ---

type grammarDef struct {
	name     string
	slug     string
	dir      string
	heading  string
	cardDesc string
	desc     string // HTML
}

var formatDefs = []grammarDef{
	{
		name:     "Beancount",
		slug:     "beancount",
		dir:      "../tree-sitter-beancount",
		heading:  "Double-entry bookkeeping language",
		cardDesc: "Double-entry bookkeeping language with account directives, multi-currency support, and validation.",
		desc: `<a href="https://beancount.github.io/">Beancount</a> is a double-entry bookkeeping language
that includes optional account directives, multi-currency support, and built-in validation.
Transactions use two or more posting legs with implicit balancing.
See also: <a href="https://www.bytestone.uk/afp/plain-text-accounting/journalasplaintext/">Journal entries as plain text</a>,
<a href="https://www.bytestone.uk/posts/abnf-and-plain-text-accounting/">ABNF and Plain Text Accounting</a>.`,
	},
	// Goluca is now sourced from docs/goluca.md via md2html.
	// See docs/goluca.md for the goluca format documentation.
	{
		name:     "PTA",
		slug:     "pta",
		dir:      "../tree-sitter-pta",
		heading:  "General plain text accounting grammar",
		cardDesc: "General-purpose plain text accounting grammar covering transactions, balance checks, and metadata.",
		desc: `PTA is a general-purpose <a href="https://plaintextaccounting.org/">plain text accounting</a> grammar
supporting transactions, postings, balance checks, data-points, and metadata.
It provides a flexible foundation that can represent entries from Ledger, hledger, and Beancount.
See also: <a href="https://www.bytestone.uk/posts/abnf/">Augmented Backus Naur Format</a>,
<a href="https://www.bytestone.uk/posts/abnf-and-plain-text-accounting/">ABNF and Plain Text Accounting</a>.`,
	},
}

// --- Helpers ---

func loadGrammarFiles(dir string) (js, jsonStr string) {
	if b, err := os.ReadFile(filepath.Join(dir, "grammar.js")); err == nil {
		js = string(b)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "src", "grammar.json")); err == nil {
		jsonStr = string(b)
	}
	return
}

func runABNF(dir string) (abnf, roundTrip string) {
	jsonPath := filepath.Join(dir, "src", "grammar.json")

	cmd := exec.Command("tree-sitter2abnf", jsonPath)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("Error: %v", err), ""
	}
	abnf = buf.String()

	tmp, err := os.CreateTemp("", "grammar-*.abnf")
	if err != nil {
		return abnf, fmt.Sprintf("Error: %v", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString(abnf); err != nil {
		return abnf, fmt.Sprintf("Error: %v", err)
	}
	if err := tmp.Close(); err != nil {
		return abnf, fmt.Sprintf("Error: %v", err)
	}

	cmd2 := exec.Command("tree-sitter2abnf", tmp.Name())
	var buf2 bytes.Buffer
	cmd2.Stdout = &buf2
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		return abnf, fmt.Sprintf("Error: %v", err)
	}
	roundTrip = buf2.String()
	return
}

func generateSVG(golucaSrc []byte) (template.HTML, error) {
	tmp, err := os.CreateTemp("", "demo-*.goluca")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(golucaSrc); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	cmd := exec.Command("pta2svg", tmp.Name())
	var svgBuf, errBuf bytes.Buffer
	cmd.Stdout = &svgBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%v: %s", err, errBuf.String())
	}
	return template.HTML(svgBuf.String()), nil
}

func buildDemos() ([]demoData, error) {
	files, err := filepath.Glob("testdata/*.beancount")
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	var demos []demoData
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		stripped := beancount.StripBlankLines(src)

		tree, err := beancount.Parse(src)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}

		bcRanges, err := beancount.Highlight(src)
		if err != nil {
			return nil, fmt.Errorf("highlight %s: %w", path, err)
		}

		golucaSrc, err := beancount.Convert(src)
		if err != nil {
			return nil, fmt.Errorf("convert %s: %w", path, err)
		}

		golucaTree, err := beancount.GolucaParse(golucaSrc)
		golucaOK := err == nil && golucaTree != nil && !golucaTree.RootNode().HasError()

		var golucaHL template.HTML
		if ranges, err := beancount.GolucaHighlight(golucaSrc); err == nil {
			golucaHL = template.HTML(renderHTML(golucaSrc, ranges))
		} else {
			golucaHL = template.HTML(html.EscapeString(string(golucaSrc)))
		}

		var svg template.HTML
		if s, err := generateSVG(golucaSrc); err != nil {
			fmt.Fprintf(os.Stderr, "svg %s: %v (skipping)\n", path, err)
		} else {
			svg = s
		}

		demos = append(demos, demoData{
			Name:               filepath.Base(path),
			BeancountHighlight: template.HTML(renderHTML(stripped, bcRanges)),
			BeancountOK:        !beancount.HasErrors(tree),
			GolucaHighlight:    golucaHL,
			GolucaOK:           golucaOK,
			SExpr:              beancount.SExpression(tree),
			SVGContent:         svg,
		})
	}
	return demos, nil
}

func writePage(tmplStr, path string, data any) error {
	t, err := template.New("page").Parse(tmplStr)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	execErr := t.Execute(f, data)
	closeErr := f.Close()
	if execErr != nil {
		return execErr
	}
	return closeErr
}

// injectGolucaABNF runs tree-sitter2abnf on the goluca grammar and
// replaces marker comments in the HTML with the highlighted output.
func injectGolucaABNF(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	golucaDir := "../tree-sitter-goluca"
	abnf, roundTrip := runABNF(golucaDir)

	abnfHTML := "<pre>" + string(highlightWithTS(".yabnf", abnf, nil)) + "</pre>"
	content = strings.Replace(content, "<!-- GENERATED:GOLUCA_ABNF -->", abnfHTML, 1)

	rtHTML := `<details><summary>ABNF → JSON round-trip</summary><pre>` +
		string(highlightWithTS(".json", roundTrip, nil)) +
		`</pre></details>`
	content = strings.Replace(content, "<!-- GENERATED:GOLUCA_ROUNDTRIP -->", rtHTML, 1)

	return os.WriteFile(path, []byte(content), 0o644)
}

// --- Main ---

// highlightCodeBlocks reads an HTML file produced by md2html, finds
// <pre><code class="language-X"> blocks, applies syntax highlighting
// (tree-sitter for goluca; regex for abnf), injects the hl-* CSS,
// and writes the file back.
func highlightCodeBlocks(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// Map of language class → highlighter function.
	type hlFunc func(string) string
	langs := map[string]hlFunc{
		"language-abnf": func(src string) string {
			return string(highlightWithTS(".yabnf", src, nil))
		},
		"language-goluca": func(src string) string {
			return string(highlightWithTS(".goluca", src, nil))
		},
	}

	re := regexp.MustCompile(`<pre><code class="(language-\w+)">([\s\S]*?)</code></pre>`)
	changed := false
	result := re.ReplaceAllStringFunc(content, func(match string) string {
		m := re.FindStringSubmatch(match)
		if len(m) < 3 {
			return match
		}
		lang, body := m[1], m[2]
		fn, ok := langs[lang]
		if !ok {
			return match
		}
		// Unescape HTML entities that goldmark produced.
		body = strings.ReplaceAll(body, "&amp;", "&")
		body = strings.ReplaceAll(body, "&lt;", "<")
		body = strings.ReplaceAll(body, "&gt;", ">")
		body = strings.ReplaceAll(body, "&quot;", `"`)
		body = strings.ReplaceAll(body, "&#39;", "'")
		changed = true
		return "<pre>" + fn(body) + "</pre>"
	})

	if !changed {
		return nil
	}

	// Inject hl-* CSS if not already present.
	if !strings.Contains(result, ".hl-keyword") {
		cssBlock := "<style>" + sharedCSS + "</style>\n</head>"
		result = strings.Replace(result, "</head>", cssBlock, 1)
	}

	return os.WriteFile(path, []byte(result), 0o644)
}

func main() {
	demos, err := buildDemos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "demos: %v\n", err)
		os.Exit(1)
	}

	// Post-process md2html-generated pages with syntax highlighting.
	for _, page := range []string{"docs/goluca.html"} {
		if err := highlightCodeBlocks(page); err != nil {
			fmt.Fprintf(os.Stderr, "highlight %s: %v\n", page, err)
			os.Exit(1)
		}
		fmt.Printf("highlighted %s\n", page)
	}

	// Inject auto-generated ABNF into goluca.html.
	if err := injectGolucaABNF("docs/goluca.html"); err != nil {
		fmt.Fprintf(os.Stderr, "goluca abnf: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("injected goluca ABNF")

	// Generate demo.html (parser demo page)
	if err := writePage(demoTmpl, "docs/demo.html", demos); err != nil {
		fmt.Fprintf(os.Stderr, "demo: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("wrote docs/demo.html")

	// Generate per-format pages
	for _, g := range formatDefs {
		js, jsonStr := loadGrammarFiles(g.dir)
		abnf, rt := runABNF(g.dir)

		page := formatPage{
			Name:        g.name,
			Slug:        g.slug,
			Heading:     g.heading,
			Description: template.HTML(g.desc),
			CardDesc:    g.cardDesc,
			GrammarJS:   highlightWithTS(".js", js, jsRules),
			GrammarJSON: highlightWithTS(".json", jsonStr, nil),
			ABNF:        highlightWithTS(".abnf", abnf, nil),
			RoundTrip:   highlightWithTS(".json", rt, nil),
		}
		if g.slug == "beancount" {
			page.Demos = demos
		}

		path := fmt.Sprintf("docs/%s.html", page.Slug)
		if err := writePage(formatTmpl, path, page); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", path)
	}
}

// --- Templates ---

const sharedCSS = `
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

// demoTmpl generates docs/demo.html — beancount→goluca conversion for each testdata file.
var demoTmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Parser Demo — Plain Text Accounting Formats</title>
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
</div>
</section>
</body>
</html>
`

// formatTmpl generates docs/{format}.html — grammar, ABNF, and demos for each format.
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

</div>
</section>
</body>
</html>
`
