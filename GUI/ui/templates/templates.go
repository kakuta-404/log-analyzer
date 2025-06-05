package templates

import (
	"GUI/internal/models"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin/render"
)

//go:embed *.gohtml
var templateFS embed.FS

type Templates struct {
	templates map[string]*template.Template
}

// Instance returns a Templates instance
//
//	func Instance() (*Templates, error) {
//		templates := make(map[string]*template.Template)
//
//		files, err := templateFS.ReadDir(".")
//		if err != nil {
//			return nil, err
//		}
//
//		for _, file := range files {
//			if file.Name() == "base.gohtml" {
//				continue
//			}
//			tmpl, err := template.ParseFS(templateFS, file.Name())
//			if err != nil {
//				return nil, err
//			}
//			templates[file.Name()] = tmpl
//		}
//
//		return &Templates{templates: templates}, nil
//	}
func Instance() (*Templates, error) {
	funcMap := template.FuncMap{
		"inc": func(i int) int { return i + 1 },
		"dec": func(i int) int { return i - 1 },
		"prevURL": func(data any) string {
			return buildNavURL(data.(map[string]any), -1)
		},
		"nextURL": func(data any) string {
			return buildNavURL(data.(map[string]any), 1)
		},
	}

	templates := make(map[string]*template.Template)

	files, err := templateFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".gohtml") || file.Name() == "base.gohtml" {
			continue
		}

		// Apply FuncMap here
		tmpl := template.New(file.Name()).Funcs(funcMap)
		tmpl, err = tmpl.ParseFS(templateFS, file.Name())
		if err != nil {
			return nil, err
		}

		templates[file.Name()] = tmpl
	}

	return &Templates{templates: templates}, nil
}

func buildNavURL(data map[string]any, delta int) string {
	projectID := data["ProjectID"].(string)
	event := data["Event"].(models.Event)
	name := event.Name
	index := data["Index"].(int) + delta
	filters := data["Filters"].(map[string]string)

	url := fmt.Sprintf("/search/detail?project_id=%s&name=%s&index=%d", projectID, name, index)
	for k, v := range filters {
		url += fmt.Sprintf("&%s=%s", k, v)
	}
	return url
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
