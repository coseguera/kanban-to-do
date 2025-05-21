// Copyright (c) 2025 Carlos Oseguera (@coseguera)
// This code is licensed under a dual-license model.
// See LICENSE.md for more information.

// Package templates handles the HTML templates used by the application
package templates

import (
	"html/template"
	"os"
	"path/filepath"
)

var (
	// Templates are the parsed HTML templates
	Templates map[string]*template.Template
)

// LoadTemplates loads and parses all HTML templates
func LoadTemplates(templatesDir string) error {
	// Create templates directory if it doesn't exist
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}

	// Parse the templates
	Templates = make(map[string]*template.Template)

	homeTmpl, err := template.ParseFiles(filepath.Join(templatesDir, "home.html"))
	if err != nil {
		return err
	}
	Templates["home"] = homeTmpl

	todoListsTmpl, err := template.ParseFiles(filepath.Join(templatesDir, "todoLists.html"))
	if err != nil {
		return err
	}
	Templates["todoLists"] = todoListsTmpl

	tasksTmpl, err := template.ParseFiles(filepath.Join(templatesDir, "tasks.html"))
	if err != nil {
		return err
	}
	Templates["tasks"] = tasksTmpl

	return nil
}
