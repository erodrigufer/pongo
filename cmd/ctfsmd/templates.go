package main

import (
	"html/template"
	"path/filepath"

	monitor "github.com/erodrigufer/CTForchestrator/internal/APIMonitor"
)

// templateData, holds the dynamic data passed to the HTML templates.
type templateData struct {
	// CurrentYear, displays the current year in the footer of the page.
	CurrentYear int
	// Username, of an SSH session.
	Username string
	// Password, of an SSH session.
	Password string
	// Port, port to which to connect with SSH session
	Port string
	// LifetimeSess, is the lifetime of a session in minutes.
	LifetimeSess int
	// OutboundIP, the oubound IP used by the host.
	OutboundIP string
	// HealthCheckResults, a slice with all the results provided by the health
	// monitor.
	HealthCheckResults []monitor.HealthCheckResult
}

// newTemplateCache, create a templates cache from a directory dir.
// The template cache stores the rendered HTML pages in a map.
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache
	cache := map[string]*template.Template{}
	// Use the filepath.Glob function to get a slice of all filepaths with
	// the extension '.page.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}
	// Loop through the pages one-by-one
	// pages is a string slice with all the *.page.tmpl files found in a
	// specific directory
	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl')
		// and assign it to the name variable
		name := filepath.Base(page)
		// First create an empty template set with template.New()
		// Finally parse the page template file into a template set.
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return nil, err
		}
		// Use the ParseGlob method to add any 'layout' templates to the
		// template set.
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}
		// Use the ParseGlob method to add any 'partial' templates to the
		// template set.
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}
		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key.
		cache[name] = ts
	}
	// Return the template set's map
	return cache, nil
}
