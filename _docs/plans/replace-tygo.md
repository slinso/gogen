# Replace tygo with gogen for knowapi.d.ts generation

## Context

You want to replace tygo with your own `gogen` tool for generating `frontend/src/lib/knowapi.d.ts` from Go source files (`requesttypes.go`, `responsetypes.go`, `dbtypes.go`). The `tygo-fix.sh` script already handles post-processing. gogen is a template-based code generator using Go AST.

## Current State of gogen

gogen already has a working TypeScript template and handles: basic structs, type aliases/named types, pointers (nullable), slices, doc comments, field tags, type mappings via config, optional fields.

## Feature Gaps (must be added to gogen first)

### Critical - Required to match tygo output

| # | Feature | Why Needed | Example from source |
|---|---------|-----------|-------------------|
| 1 | **Multi-file input** | tygo processes 3 files into one output | `-i requesttypes.go,responsetypes.go,dbtypes.go` |
| 2 | **Embedded struct → `extends`** | Response types use `tstype:",extends"` tag | `SensorMeasureResponse struct { SensorMeasure \`tstype:",extends"\` ... }` → `export interface SensorMeasureResponse extends SensorMeasure { ... }` |
| 3 | **Constants** | FilterOperation consts need TS `export const` | `const FilterOpEq FilterOperation = "eq"` → `export const FilterOpEq: FilterOperation = 'eq';` |
| 4 | **Map types** | Several fields use `map[string]T` | `Fields map[string]ColumnMapping` → `{ [key: string]: ColumnMapping }` |
| 5 | **Type filtering (per-file regex)** | dbtypes.go only exports ~20 of hundreds of types | `include_patterns: ["dbtypes.go:Catalog$", ...]` |
| 6 | **Go type comments with `/* int */` annotations** | tygo preserves Go type info as comments | `id: number /* int */;` |

### Nice-to-have (can work around)

| # | Feature | Notes |
|---|---------|-------|
| 7 | **Generic types** | Only `SingleValue[T comparable]` — could hardcode in template or type_mappings |
| 8 | **Anonymous inline struct slices** | Only `ExcelMappingMessage` — could restructure Go code |
| 9 | **Frontmatter/imports** | tygo has `frontmatter` config — gogen could add header config or template it |

## Recommendation

**You should move to the gogen project first and add features before attempting the switch.**

The critical gaps (especially #1 multi-file, #2 extends, #3 constants, #4 maps) are foundational features that can't be worked around with template tricks. Here's a suggested order:

### Implementation Order for gogen

1. **Map type support** — Add `{ [key: string]: ValueType }` generation (parser likely already handles maps, just needs template function)
2. **Constants** — Parse `const` blocks, expose them in template data
3. **Embedded struct extends** — Detect `tstype:",extends"` tag, add `extends` template helper
4. **Multi-file input** — Accept multiple `-i` files, merge types
5. **Type filtering** — Add include/exclude patterns per file
6. **Go type comment annotations** — Add option to include `/* int */` style comments

### After gogen features are ready

Update the typescript.tmpl to produce output matching current knowapi.d.ts, then:
- Update `tygo-fix.sh` → `gogen-gen.sh`
- Remove tygo dependency
- Single invocation: `gogen -i requesttypes.go,responsetypes.go,dbtypes.go -t typescript.tmpl -c gogen.yaml -o frontend/src/lib/knowapi.d.ts`

## Verification

- Generate output with both tygo and gogen, diff them
- Run `pnpm check` in frontend/ to verify TypeScript compatibility
