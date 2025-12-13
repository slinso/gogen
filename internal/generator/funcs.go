package generator

import (
	"fmt"
	"strings"
	"text/template"
	"unicode"

	"gogen/internal/config"
	"gogen/internal/model"
)

// ValidateRule represents a parsed validation rule from a validate tag.
type ValidateRule struct {
	Name  string // Rule name (e.g., "min", "max", "email")
	Value string // Rule value (e.g., "1", "45", empty for boolean rules)
}

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

		// Valibot form helpers
		"valibotFormField": valibotFormField,
		"hasValidateRule":  hasValidateRule,
		"getValidateValue": getValidateValue,
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

// parseValidateTag parses a validate struct tag and returns a list of rules.
// Input: "required,min=1,max=45"
// Output: []ValidateRule{{Name: "required"}, {Name: "min", Value: "1"}, {Name: "max", Value: "45"}}
func parseValidateTag(field model.Field) []ValidateRule {
	tagValue, ok := field.Tag.Values["validate"]
	if !ok || tagValue == "" {
		return nil
	}

	var rules []ValidateRule
	parts := strings.Split(tagValue, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if idx := strings.Index(part, "="); idx > 0 {
			rules = append(rules, ValidateRule{
				Name:  part[:idx],
				Value: part[idx+1:],
			})
		} else {
			rules = append(rules, ValidateRule{
				Name: part,
			})
		}
	}
	return rules
}

// hasValidateRule checks if a field has a specific validate rule.
func hasValidateRule(field model.Field, ruleName string) bool {
	for _, rule := range parseValidateTag(field) {
		if rule.Name == ruleName {
			return true
		}
	}
	return false
}

// getValidateValue gets the value for a validate rule.
func getValidateValue(field model.Field, ruleName string) string {
	for _, rule := range parseValidateTag(field) {
		if rule.Name == ruleName {
			return rule.Value
		}
	}
	return ""
}

// valibotFormField generates a Valibot field expression for form validation.
// It uses v.optional() with defaults and adds validators from validate tags.
func valibotFormField(field model.Field) string {
	typeKind := field.Type.Kind
	typeName := field.Type.Name

	// Determine base type and default value
	var baseType, defaultVal string
	isNumeric := false

	switch typeKind {
	case model.KindBasic:
		switch {
		case typeName == "string":
			baseType = "v.string()"
			defaultVal = "''"
		case typeName == "bool":
			baseType = "v.boolean()"
			defaultVal = "false"
		case strings.HasPrefix(typeName, "int") || strings.HasPrefix(typeName, "uint") ||
			strings.HasPrefix(typeName, "float") || typeName == "byte" || typeName == "rune":
			baseType = "v.number()"
			defaultVal = "0"
			isNumeric = true
		default:
			baseType = "v.unknown()"
			defaultVal = "undefined"
		}
	case model.KindNamed:
		// Handle special named types
		if field.Type.Raw == "time.Time" {
			baseType = "v.string()"
			defaultVal = "''"
		} else if field.Type.Raw == "uuid.UUID" {
			baseType = "v.string()"
			defaultVal = "''"
		} else {
			// Reference to another schema
			return fmt.Sprintf("%sSchema", field.Type.Name)
		}
	case model.KindSlice, model.KindArray:
		elemType := valibotElemType(field.Type.Elem)
		return fmt.Sprintf("v.optional(v.array(%s), [])", elemType)
	case model.KindMap:
		keyType := valibotElemType(field.Type.Key)
		valueType := valibotElemType(field.Type.Value)
		return fmt.Sprintf("v.optional(v.record(%s, %s), {})", keyType, valueType)
	case model.KindPointer:
		// Pointers are nullable
		elemType := valibotElemType(field.Type.Elem)
		return fmt.Sprintf("v.nullable(%s)", elemType)
	default:
		baseType = "v.unknown()"
		defaultVal = "undefined"
	}

	// Build validators from validate tag
	var validators []string
	rules := parseValidateTag(field)

	for _, rule := range rules {
		switch rule.Name {
		case "min":
			if isNumeric {
				validators = append(validators, fmt.Sprintf("v.minValue(%s)", rule.Value))
			} else {
				validators = append(validators, fmt.Sprintf("v.minLength(%s)", rule.Value))
			}
		case "max":
			if isNumeric {
				validators = append(validators, fmt.Sprintf("v.maxValue(%s)", rule.Value))
			} else {
				validators = append(validators, fmt.Sprintf("v.maxLength(%s)", rule.Value))
			}
		case "email":
			validators = append(validators, "v.email()")
		case "url":
			validators = append(validators, "v.url()")
		case "uuid":
			validators = append(validators, "v.uuid()")
		}
	}

	// Build the final expression
	optionalExpr := fmt.Sprintf("v.optional(%s, %s)", baseType, defaultVal)

	if len(validators) == 0 {
		return optionalExpr
	}

	// Use v.pipe() when we have validators
	parts := append([]string{optionalExpr}, validators...)
	return fmt.Sprintf("v.pipe(%s)", strings.Join(parts, ", "))
}

// valibotElemType returns the Valibot type for a TypeRef element.
func valibotElemType(t *model.TypeRef) string {
	if t == nil {
		return "v.unknown()"
	}

	switch t.Kind {
	case model.KindBasic:
		switch t.Name {
		case "string":
			return "v.string()"
		case "bool":
			return "v.boolean()"
		default:
			if strings.HasPrefix(t.Name, "int") || strings.HasPrefix(t.Name, "uint") ||
				strings.HasPrefix(t.Name, "float") || t.Name == "byte" || t.Name == "rune" {
				return "v.number()"
			}
			return "v.unknown()"
		}
	case model.KindNamed:
		if t.Raw == "time.Time" {
			return "v.pipe(v.string(), v.isoDateTime())"
		} else if t.Raw == "uuid.UUID" {
			return "v.pipe(v.string(), v.uuid())"
		}
		return fmt.Sprintf("%sSchema", t.Name)
	case model.KindSlice, model.KindArray:
		return fmt.Sprintf("v.array(%s)", valibotElemType(t.Elem))
	case model.KindMap:
		return fmt.Sprintf("v.record(%s, %s)", valibotElemType(t.Key), valibotElemType(t.Value))
	case model.KindPointer:
		return fmt.Sprintf("v.nullable(%s)", valibotElemType(t.Elem))
	case model.KindInterface:
		return "v.unknown()"
	default:
		return "v.unknown()"
	}
}
