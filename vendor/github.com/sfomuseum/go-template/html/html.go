// Package html provides methods for loading HTML (.html) templates with default functions
package html

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/sfomuseum/go-template/funcs"
)

// LoadTemplates loads HTML (.html) from 't_fs' with default functions assigned.
func LoadTemplates(ctx context.Context, t_fs ...fs.FS) (*template.Template, error) {

	funcs := TemplatesFuncMap()
	excluding := make([]string, 0)

	return LoadTemplatesWithFuncMapExcluding(ctx, funcs, excluding, t_fs...)
}

// LoadTemplatesWithFuncs loads HTML (.html) from 't_fs' with template functions defined by 'funcs'.
func LoadTemplatesWithFuncMap(ctx context.Context, funcs template.FuncMap, t_fs ...fs.FS) (*template.Template, error) {

	excluding := make([]string, 0)
	return LoadTemplatesWithFuncMapExcluding(ctx, funcs, excluding, t_fs...)
}

// LoadTemplatesExcluding loads HTML (.html) from 't_fs' with default functions assigned excluding
// templates with (template) names matching 'exclude_list'.
func LoadTemplatesExcluding(ctx context.Context, excluding []string, t_fs ...fs.FS) (*template.Template, error) {

	funcs := TemplatesFuncMap()
	return LoadTemplatesWithFuncMapExcluding(ctx, funcs, excluding, t_fs...)
}

// LoadTemplatesWithFuncMapExcluding loads HTML (.html) from 't_fs' with functions defined by 'funcs' excluding
// templates with (template) names matching 'exclude_list'.
func LoadTemplatesWithFuncMapExcluding(ctx context.Context, funcs template.FuncMap, excluding []string, t_fs ...fs.FS) (*template.Template, error) {

	t := template.New("html").Funcs(funcs)

	for idx, f := range t_fs {

		var err error

		if len(excluding) == 0 {

			t, err = t.ParseFS(f, "*.html")

			if err != nil {
				return nil, fmt.Errorf("Failed to load templates from FS at offset %d, %w", idx, err)
			}

			continue
		}

		// filter templates here

		err = fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {

			if err != nil {
				return fmt.Errorf("Encountered an error walking %s, %w", path, err)
			}

			if d.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".html" {
				return nil
			}

			r, err := f.Open(path)

			if err != nil {
				return fmt.Errorf("Failed to open %s for reading, %w", path, err)
			}

			defer r.Close()

			body, err := io.ReadAll(r)

			if err != nil {
				return fmt.Errorf("Failed to read %s, %w", path, err)
			}

			tmp_t, err := template.New("__tmp__").Funcs(funcs).Parse(string(body))

			if err != nil {
				return fmt.Errorf("Failed to parse %s, %w", path, err)
			}

			exclude_t := false

			for _, test_t := range tmp_t.Templates() {

				t_name := test_t.Name()

				if t_name == "__tmp__" {
					continue
				}

				for _, name := range excluding {

					if t_name == name {
						exclude_t = true
						break
					}
				}

				if exclude_t {
					break
				}
			}

			if !exclude_t {
				t, _ = t.Parse(string(body))
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("Failed to parse templates, %w", err)
		}

	}

	return t, nil
}

// TemplatesFuncMap() returns a `template.FuncMap` instance with default functions assigned.
func TemplatesFuncMap() template.FuncMap {

	return template.FuncMap{
		// For example: {{ if (IsAvailable "Account" .) }}
		"IsAvailable":      funcs.IsAvailable,
		"Add":              funcs.Add,
		"JoinPath":         funcs.JoinPath,
		"QRCodeB64":        funcs.QRCodeB64,
		"QRCodeDataURI":    funcs.QRCodeDataURI,
		"IsEven":           funcs.IsEven,
		"IsOdd":            funcs.IsOdd,
		"FormatStringTime": funcs.FormatStringTime,
		"FormatUnixTime":   funcs.FormatUnixTime,
		"GjsonGet":         funcs.GjsonGet,
		"StringHasPrefix":  funcs.StringHasPrefix,
		"URLQueryEscape":   funcs.URLQueryEscape,
	}
}
