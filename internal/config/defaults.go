// Package config provides configuration handling for gogen.
package config

// DefaultTypeMappings returns default Go to TypeScript type mappings.
func DefaultTypeMappings() map[string]string {
	return map[string]string{
		// Basic types
		"string":     "string",
		"bool":       "boolean",
		"int":        "number",
		"int8":       "number",
		"int16":      "number",
		"int32":      "number",
		"int64":      "number",
		"uint":       "number",
		"uint8":      "number",
		"uint16":     "number",
		"uint32":     "number",
		"uint64":     "number",
		"float32":    "number",
		"float64":    "number",
		"complex64":  "number",
		"complex128": "number",
		"byte":       "number",
		"rune":       "number",
		"uintptr":    "number",

		// Special types
		"[]byte":        "string", // Often base64 encoded
		"time.Time":     "string", // ISO 8601 string
		"time.Duration": "number", // Nanoseconds as number
		"interface{}":   "unknown",
		"any":           "unknown",
		"error":         "string", // Error message

		// UUID types (common libraries)
		"uuid.UUID":                       "string",
		"github.com/google/uuid.UUID":     "string",
		"github.com/gofrs/uuid.UUID":      "string",
		"github.com/satori/go.uuid.UUID":  "string",

		// Decimal types
		"decimal.Decimal":                     "string",
		"github.com/shopspring/decimal.Decimal": "string",

		// JSON types
		"json.RawMessage": "unknown",
	}
}

// DefaultOptions returns default generation options.
func DefaultOptions() Options {
	return Options{
		PerType:      false,
		ExportedOnly: true,
		TagKey:       "json",
	}
}
