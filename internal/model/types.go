// Package model defines the intermediate representation for parsed Go types.
package model

// TypeKind represents the category of a Go type.
type TypeKind string

const (
	KindStruct    TypeKind = "struct"
	KindAlias     TypeKind = "alias"
	KindNamed     TypeKind = "named"
	KindBasic     TypeKind = "basic"
	KindSlice     TypeKind = "slice"
	KindArray     TypeKind = "array"
	KindMap       TypeKind = "map"
	KindPointer   TypeKind = "pointer"
	KindInterface TypeKind = "interface"
)

// File represents a parsed Go source file.
type File struct {
	Package string   // Package name
	Path    string   // File path
	Types   []Type   // All type definitions
	Imports []Import // Import statements
}

// Import represents a Go import statement.
type Import struct {
	Alias string // Optional alias (empty if none)
	Path  string // Import path
}

// Type represents a Go type definition.
type Type struct {
	Name       string   // Type name (e.g., "User")
	Kind       TypeKind // Type category
	Doc        string   // Documentation comment
	Fields     []Field  // Fields (for structs)
	Underlying *TypeRef // Underlying type (for aliases/named types)
	IsExported bool     // Whether the type is exported
}

// Field represents a struct field.
type Field struct {
	Name       string    // Field name (empty for embedded)
	Type       TypeRef   // Field type reference
	Tag        StructTag // Struct tag
	Doc        string    // Documentation comment
	IsExported bool      // Whether the field is exported
	IsEmbedded bool      // Whether this is an embedded field
}

// TypeRef represents a reference to a type.
type TypeRef struct {
	Kind    TypeKind // Type category
	Name    string   // Type name (for named/basic types)
	Package string   // Package name (for imported types, e.g., "time" for time.Time)
	Elem    *TypeRef // Element type (for slice, array, pointer)
	Key     *TypeRef // Key type (for maps)
	Value   *TypeRef // Value type (for maps)
	Raw     string   // Raw Go type string representation
}

// StructTag represents parsed struct tags.
type StructTag struct {
	Raw    string            // Raw tag string
	Values map[string]string // Parsed tag values (key -> value)
}

// FullName returns the full qualified name of a TypeRef (e.g., "time.Time").
func (t *TypeRef) FullName() string {
	if t.Package != "" {
		return t.Package + "." + t.Name
	}
	return t.Name
}
