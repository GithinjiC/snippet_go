package main

import (
	"cosmasgithinji.net/simplesnippetbox/pkg/forms"
	"cosmasgithinji.net/simplesnippetbox/pkg/models"
	"html/template"
	// "net/url"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"path/filepath"
	"strings"
	"time"
)

// dynamic data passed to HTML templates
type templateData struct {
	AuthenticatedUser *models.User
	CSRFToken         string
	CurrentYear       int
	Flash             string
	Form              *forms.Form
	Snippet           *models.Snippet
	Snippets          []*models.Snippet
}

// template functions
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:06")
}

func capitalize(name string) string {
	titleCaser := cases.Title(language.English) // Initialize the title caser
	return titleCaser.String(strings.ToLower(name))
}

var functions = template.FuncMap{
	"humanDate":  humanDate,
	"capitalize": capitalize,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{} // init new map to act as cache

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.go.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		//add layout templates
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.go.tmpl"))
		if err != nil {
			return nil, err
		}
		//add partial templates
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.go.tmpl"))
		if err != nil {
			return nil, err
		}
		cache[name] = ts //Add template set to cache
	}
	return cache, nil //Return the map
}
