package prompt_test

import (
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/llm/prompt"
)

func TestTemplateRender(t *testing.T) {
	tmpl := &prompt.Template{
		Name:   "test",
		System: "You are {{.Role}}",
		User:   "Summarize: {{.Text}}",
	}
	sys, user := tmpl.Render(map[string]any{"Role": "assistant", "Text": "hello"})
	if sys != "You are assistant" || user != "Summarize: hello" {
		t.Fatalf("sys=%q user=%q", sys, user)
	}
}

func TestBuiltinTemplates(t *testing.T) {
	cases := []struct {
		tmpl *prompt.Template
		vars map[string]any
		want string
	}{
		{prompt.Summarize, map[string]any{"Text": "long article"}, "long article"},
		{prompt.Translate, map[string]any{"SourceLang": "en", "TargetLang": "zh", "Text": "hi"}, "hi"},
		{prompt.QA, map[string]any{"Context": "doc", "Question": "what?"}, "what?"},
	}
	for _, c := range cases {
		_, user := c.tmpl.Render(c.vars)
		if !strings.Contains(user, c.want) {
			t.Fatalf("%s user=%q", c.tmpl.Name, user)
		}
	}
}
