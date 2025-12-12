// Package generator provides template-based code generation.
package generator

import (
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"gogen/internal/config"
	"gogen/internal/model"
)

// Generator executes templates against parsed types.
type Generator struct {
	config   *config.Config
	template *template.Template
}

// New creates a new Generator.
func New(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
	}
}

// LoadTemplate loads a template from file.
func (g *Generator) LoadTemplate(path string) error {
	tmpl, err := template.New(filepath.Base(path)).
		Funcs(templateFuncs(g.config)).
		ParseFiles(path)
	if err != nil {
		return fmt.Errorf("loading template: %w", err)
	}
	g.template = tmpl
	return nil
}

// TemplateData represents data passed to templates.
type TemplateData struct {
	File         *model.File       // The parsed file
	Types        []model.Type      // Types to generate (filtered)
	Type         *model.Type       // Current type (for per-type mode)
	Config       *config.Config    // Configuration
	TypeMappings map[string]string // Type mappings for convenience
}

// Generate generates output for all types.
func (g *Generator) Generate(file *model.File, w io.Writer) error {
	types := g.filterTypes(file.Types)

	// Build type map for embedded field flattening
	typeMap := make(map[string]model.Type)
	for _, t := range file.Types {
		typeMap[t.Name] = t
	}

	// Flatten embedded fields
	types = g.flattenEmbedded(types, typeMap)

	if g.config.Options.PerType {
		// Execute template once per type
		for i := range types {
			data := &TemplateData{
				File:         file,
				Types:        types,
				Type:         &types[i],
				Config:       g.config,
				TypeMappings: g.config.TypeMappings,
			}
			if err := g.template.Execute(w, data); err != nil {
				return fmt.Errorf("executing template for %s: %w", types[i].Name, err)
			}
		}
	} else {
		// Execute template once for all types
		data := &TemplateData{
			File:         file,
			Types:        types,
			Config:       g.config,
			TypeMappings: g.config.TypeMappings,
		}
		if err := g.template.Execute(w, data); err != nil {
			return fmt.Errorf("executing template: %w", err)
		}
	}

	return nil
}

// filterTypes filters types based on configuration.
func (g *Generator) filterTypes(types []model.Type) []model.Type {
	var result []model.Type

	for _, t := range types {
		if g.config.ShouldIncludeType(t.Name, t.IsExported) {
			result = append(result, t)
		}
	}

	return result
}

// flattenEmbedded flattens embedded fields into their parent structs.
func (g *Generator) flattenEmbedded(types []model.Type, typeMap map[string]model.Type) []model.Type {
	result := make([]model.Type, 0, len(types))

	for _, t := range types {
		if t.Kind != model.KindStruct {
			result = append(result, t)
			continue
		}

		flattened := g.flattenFields(t.Fields, typeMap, make(map[string]bool))
		t.Fields = flattened
		result = append(result, t)
	}

	return result
}

// flattenFields recursively flattens embedded fields.
func (g *Generator) flattenFields(fields []model.Field, typeMap map[string]model.Type, seen map[string]bool) []model.Field {
	var result []model.Field

	for _, f := range fields {
		if !f.IsEmbedded {
			result = append(result, f)
			continue
		}

		// Get the embedded type name
		typeName := f.Type.Name
		if typeName == "" && f.Type.Elem != nil {
			// Pointer to embedded type
			typeName = f.Type.Elem.Name
		}

		// Prevent infinite recursion
		if seen[typeName] {
			continue
		}
		seen[typeName] = true

		// Look up the embedded type
		embeddedType, ok := typeMap[typeName]
		if !ok || embeddedType.Kind != model.KindStruct {
			// If we can't find it or it's not a struct, skip it
			// (it might be from an external package)
			continue
		}

		// Recursively flatten the embedded type's fields
		embeddedFields := g.flattenFields(embeddedType.Fields, typeMap, seen)
		result = append(result, embeddedFields...)

		delete(seen, typeName)
	}

	return result
}
