// gogen is a Go type code generator that parses Go source files
// and generates code using templates.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gogen/internal/config"
	"gogen/internal/generator"
	"gogen/internal/parser"
)

var (
	inputFile    string
	templateFile string
	configFile   string
	outputFile   string
	perType      bool
	exportedOnly bool
	tagKey       string
	types        string
	exclude      string
	verbose      bool
	showHelp     bool
)

func init() {
	flag.StringVar(&inputFile, "input", "", "Input Go source file (required)")
	flag.StringVar(&inputFile, "i", "", "Input Go source file (shorthand)")

	flag.StringVar(&templateFile, "template", "", "Template file (required)")
	flag.StringVar(&templateFile, "t", "", "Template file (shorthand)")

	flag.StringVar(&configFile, "config", "", "Config file (YAML/JSON)")
	flag.StringVar(&configFile, "c", "", "Config file (shorthand)")

	flag.StringVar(&outputFile, "output", "", "Output file (default: stdout)")
	flag.StringVar(&outputFile, "o", "", "Output file (shorthand)")

	flag.BoolVar(&perType, "per-type", false, "Execute template once per type")
	flag.BoolVar(&exportedOnly, "exported", true, "Only process exported types")
	flag.StringVar(&tagKey, "tag", "json", "Tag key for field names")
	flag.StringVar(&types, "types", "", "Only generate for these types (comma-separated)")
	flag.StringVar(&types, "T", "", "Only generate for these types (shorthand)")
	flag.StringVar(&exclude, "exclude", "", "Exclude these types (comma-separated)")
	flag.StringVar(&exclude, "X", "", "Exclude these types (shorthand)")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `gogen - Go type code generator

Usage:
    gogen -i <input.go> -t <template.tmpl> [options]

Options:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
    # Generate TypeScript types
    gogen -i models.go -t typescript.tmpl -o models.ts

    # Generate schema for specific structs only
    gogen -i models.go -t zod.tmpl -T User,Product -o schemas.ts

    # Generate with multiple types (alternative syntax)
    gogen -i models.go -t typescript.tmpl --types "User, Order, Product"

    # Exclude specific types
    gogen -i models.go -t typescript.tmpl -X InternalConfig,PrivateData

    # Generate with custom config
    gogen -i models.go -t zod.tmpl -c config.yaml -o schemas.ts

    # Generate per-type output to stdout
    gogen -i models.go -t typescript.tmpl --per-type

    # Only process specific tag
    gogen -i models.go -t typescript.tmpl --tag yaml

`)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	if showHelp {
		flag.Usage()
		return nil
	}

	// Validate required flags
	if inputFile == "" {
		return fmt.Errorf("input file is required (-i or --input)")
	}
	if templateFile == "" {
		return fmt.Errorf("template file is required (-t or --template)")
	}

	// Load configuration
	cfg := config.New()
	if configFile != "" {
		if err := cfg.LoadFile(configFile); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	}

	// Apply CLI overrides
	if perType {
		cfg.Options.PerType = true
	}
	cfg.Options.ExportedOnly = exportedOnly
	if tagKey != "" {
		cfg.Options.TagKey = tagKey
	}
	if types != "" {
		cfg.Options.IncludeTypes = parseCommaSeparated(types)
	}
	if exclude != "" {
		cfg.Options.ExcludeTypes = parseCommaSeparated(exclude)
	}

	// Parse input file
	p := parser.New()
	file, err := p.ParseFile(inputFile)
	if err != nil {
		return fmt.Errorf("parsing input: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Parsed %d types from %s\n", len(file.Types), inputFile)
		for _, t := range file.Types {
			fmt.Fprintf(os.Stderr, "  - %s (%s)\n", t.Name, t.Kind)
		}
	}

	// Create generator and load template
	gen := generator.New(cfg)
	if err := gen.LoadTemplate(templateFile); err != nil {
		return err
	}

	// Determine output destination
	var output *os.File
	if outputFile != "" {
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	// Generate output
	if err := gen.Generate(file, output); err != nil {
		return err
	}

	if verbose && outputFile != "" {
		fmt.Fprintf(os.Stderr, "Generated output to %s\n", outputFile)
	}

	return nil
}

// parseCommaSeparated splits a comma-separated string into a slice of trimmed strings.
func parseCommaSeparated(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
