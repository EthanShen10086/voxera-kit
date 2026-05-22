// Package prompt provides parameterized prompt templates for common LLM tasks.
package prompt

import (
	"bytes"
	"text/template"
)

// Template holds a parameterized prompt template with separate system and user
// portions. Variables use Go's text/template syntax (e.g. {{.VarName}}).
type Template struct {
	Name   string
	System string
	User   string
}

// Render substitutes the given variables into both the system and user
// templates, returning the rendered strings. If a template is empty or
// rendering fails, the corresponding output is an empty string.
func (t *Template) Render(vars map[string]any) (system string, user string) {
	system = renderOne(t.Name+"_system", t.System, vars)
	user = renderOne(t.Name+"_user", t.User, vars)
	return system, user
}

func renderOne(name, tmplStr string, vars map[string]any) string {
	if tmplStr == "" {
		return ""
	}
	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return ""
	}
	return buf.String()
}
