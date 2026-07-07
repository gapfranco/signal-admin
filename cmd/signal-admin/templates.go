package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"signal-admin/ui"
	"strings"
	"time"
)

type templateData struct {
	CurrentYear     int
	Flash           string
	IsAuthenticated bool
	CurrentUser     string
	CSRFToken       string
	Form            any
	Data            any
	ActiveMenu      string
	ActiveSubmenu   string
	NomeEmpresa     string
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		baseFile := "html/base.html"
		if name == "login.html" || name == "setup.html" {
			baseFile = "html/base_login.html"
		}

		patterns := []string{
			baseFile,
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

var functions = template.FuncMap{
	"userInitial": func(s string) string {
		if s == "" {
			return "?"
		}
		return strings.ToUpper(string([]rune(s)[0]))
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("parameters must be even")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"formatDateBR": func(s string) string {
		if s == "" {
			return "—"
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return s
		}
		return t.Format("02/01/2006")
	},
	"statusLabel": func(s string) string {
		switch s {
		case "active":
			return "Ativo"
		case "suspended":
			return "Suspenso"
		case "inactive":
			return "Inativo"
		default:
			return s
		}
	},
}
