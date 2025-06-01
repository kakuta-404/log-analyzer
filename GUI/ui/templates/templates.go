package templates

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin/render"
)

//go:embed *.gohtml
var templateFS embed.FS

type Templates struct {
	templates map[string]*template.Template
}

// Instance returns a Templates instance
func Instance() (*Templates, error) {
	templates := make(map[string]*template.Template)

	files, err := templateFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Name() == "base.gohtml" {
			continue
		}
		tmpl, err := template.ParseFS(templateFS, file.Name())
		if err != nil {
			return nil, err
		}
		templates[file.Name()] = tmpl
	}

	return &Templates{templates: templates}, nil
}

func (t *Templates) Instance(name string, data any) render.Render {
	return &TemplateRenderer{
		template: t.templates[name],
		name:     name,
		data:     data,
	}
}

type TemplateRenderer struct {
	template *template.Template
	name     string
	data     any
}

func (t *TemplateRenderer) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	return t.template.Execute(w, t.data)
}

func (t *TemplateRenderer) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}
