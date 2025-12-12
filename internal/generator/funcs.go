package generator

import (
	"strings"
	"text/template"
	"unicode"

	"gogen/internal/config"
	"gogen/internal/model"
)

// templateFuncs returns custom template functions.
func templateFuncs(cfg *config.Config) template.FuncMap {
	return template.FuncMap{
		// Type mapping
		"mapType": func(t model.TypeRef) string {
			return mapType(cfg, t)
		},

		// String manipulation
		"camelCase":  camelCase,
		"pascalCase": pascalCase,
		"snakeCase":  snakeCase,
		"kebabCase":  kebabCase,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"trim":       strings.TrimSpace,
		"replace":    strings.ReplaceAll,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,

		// Tag helpers
		"tag":       getTag,
		"tagOrName": func(f model.Field) string { return tagOrName(f, cfg.Options.TagKey) },
		"jsonName":  jsonName,
		"hasTag":    hasTag,

		// Type helpers
		"isStruct":    func(t model.TypeRef) bool { return t.Kind == model.KindStruct },
		"isSlice":     func(t model.TypeRef) bool { return t.Kind == model.KindSlice },
		"isArray":     func(t model.TypeRef) bool { return t.Kind == model.KindArray },
		"isMap":       func(t model.TypeRef) bool { return t.Kind == model.KindMap },
		"isPointer":   func(t model.TypeRef) bool { return t.Kind == model.KindPointer },
		"isBasic":     func(t model.TypeRef) bool { return t.Kind == model.KindBasic },
		"isInterface": func(t model.TypeRef) bool { return t.Kind == model.KindInterface },
		"isOptional":  isOptional,
		"elemType": func(t model.TypeRef) *model.TypeRef {
			return t.Elem
		},
		"keyType": func(t model.TypeRef) *model.TypeRef {
			return t.Key
		},
		"valueType": func(t model.TypeRef) *model.TypeRef {
			return t.Value
		},

		// List helpers
		"join":     strings.Join,
		"contains": containsStr,

		// Conditional helpers
		"default": defaultValue,
		"ternary": ternary,

		// Comment formatting
		"comment":    formatComment,
		"docComment": formatDocComment,

		// Misc
		"notLast": func(i, length int) bool { return i < length-1 },
	}
}

// mapType maps a Go type to the target language type.
func mapType(cfg *config.Config, t model.TypeRef) string {
	// Check for exact raw match first
	if mapped := cfg.MapType(t.Raw); mapped != t.Raw {
		return mapped
	}

	// Check for full name match (package.Type)
	if t.Package != "" {
		fullName := t.Package + "." + t.Name
		if mapped := cfg.MapType(fullName); mapped != fullName {
			return mapped
		}
	}

	// Check for basic type name match
	if mapped := cfg.MapType(t.Name); mapped != t.Name {
		return mapped
	}

	// Handle composite types
	switch t.Kind {
	case model.KindSlice, model.KindArray:
		if t.Elem != nil {
			return mapType(cfg, *t.Elem) + "[]"
		}
	case model.KindMap:
		if t.Key != nil && t.Value != nil {
			return "Record<" + mapType(cfg, *t.Key) + ", " + mapType(cfg, *t.Value) + ">"
		}
	case model.KindPointer:
		if t.Elem != nil {
			return mapType(cfg, *t.Elem) + " | null"
		}
	case model.KindInterface:
		return "unknown"
	}

	// Default: use the type name as-is
	return t.Name
}

// tagOrName returns the tag value for key, or the field name.
func tagOrName(field model.Field, key string) string {
	if val, ok := field.Tag.Values[key]; ok {
		parts := strings.Split(val, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}
	return field.Name
}

// jsonName returns the JSON field name.
func jsonName(field model.Field) string {
	return tagOrName(field, "json")
}

// getTag returns the raw tag value for a key.
func getTag(field model.Field, key string) string {
	return field.Tag.Values[key]
}

// hasTag checks if a field has a specific tag.
func hasTag(field model.Field, key string) bool {
	_, ok := field.Tag.Values[key]
	return ok
}

// isOptional checks if a field is optional (pointer or has omitempty).
func isOptional(field model.Field) bool {
	if field.Type.Kind == model.KindPointer {
		return true
	}
	if jsonTag, ok := field.Tag.Values["json"]; ok {
		return strings.Contains(jsonTag, "omitempty")
	}
	return false
}

// camelCase converts to camelCase.
func camelCase(s string) string {
	if s == "" {
		return s
	}
	pascal := pascalCase(s)
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// pascalCase converts to PascalCase.
func pascalCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, "")
}

// snakeCase converts to snake_case.
func snakeCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "_")
}

// kebabCase converts to kebab-case.
func kebabCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "-")
}

// splitWords splits a string into words (handles camelCase, PascalCase, snake_case, etc.).
func splitWords(s string) []string {
	var words []string
	var current []rune

	for i, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
			continue
		}

		if unicode.IsUpper(r) && i > 0 {
			// Check if this is the start of a new word
			prev := rune(s[i-1])
			if unicode.IsLower(prev) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1]))) {
				if len(current) > 0 {
					words = append(words, string(current))
					current = nil
				}
			}
		}

		current = append(current, r)
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}

// formatComment formats a comment with a prefix.
func formatComment(comment, prefix string) string {
	if comment == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(comment), "\n")
	var result []string
	for _, line := range lines {
		result = append(result, prefix+strings.TrimSpace(line))
	}
	return strings.Join(result, "\n")
}

// formatDocComment formats a documentation comment for TypeScript.
func formatDocComment(comment string) string {
	if comment == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(comment), "\n")
	if len(lines) == 1 {
		return "/** " + strings.TrimSpace(lines[0]) + " */"
	}
	var result []string
	result = append(result, "/**")
	for _, line := range lines {
		result = append(result, " * "+strings.TrimSpace(line))
	}
	result = append(result, " */")
	return strings.Join(result, "\n")
}

// containsStr checks if a slice contains a string.
func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// defaultValue returns the first non-empty value.
func defaultValue(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

// ternary returns a if condition is true, else b.
func ternary(condition bool, a, b string) string {
	if condition {
		return a
	}
	return b
}
