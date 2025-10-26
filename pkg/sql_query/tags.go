package sql_query

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var mu = &sync.Mutex{}

// ExtractJSONTags returns columns with near-zero cost on warm cache.
// To avoid complexity in usage, it's better to ignore params and just passing dto.SomeStruct to generic
func ExtractJSONTags[T any](optType ...reflect.Type) []string {
	var typ reflect.Type
	// If outside this function already passing value of `reflect.TypeOf()` (like an example, in benchmark tests)
	if len(optType) > 0 {
		typ = optType[0]
	} else {
		// To get type of given generic type
		typ = reflect.TypeOf((*T)(nil)).Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	// Get package path followed by struct name to give strict uniqueness to map's key
	typeName := typ.PkgPath() + "." + typ.Name()

	// Directly looking for existing columns for this struct in memory cache
	if cols, ok := columnsCache[typeName]; ok {
		return *cols
	}

	mu.Lock()
	defer mu.Unlock()
	// If columns not cached, fallback to process below
	fieldMeta, ok := fieldMetaCache[typeName]
	if !ok {
		// Build metadata if not cached (only once per type)
		meta := ExtractFromType(typ)
		fieldMetaCache[typeName] = &meta
		fieldMeta = &meta
	}

	// Build final columns (only once per type)
	// This generated columns will be saved in memory cache for next call
	cols := buildColumnsFromMeta(*fieldMeta)

	columnsCache[typeName] = cols

	return *cols
}

// Generate slice of strings based on given field meta.
// Loop through slice of field meta for 1 type/struct
func buildColumnsFromMeta(fieldMeta []FieldMeta) *[]string {
	skipped := []string{"", "-"}
	cols := new([]string)

	for i := range fieldMeta {
		meta := fieldMeta[i]
		if ArrayIncludes(skipped, meta.JSONTag) {
			continue
		}

		// Basically process below are conditions to build a proper sql expression for respective field
		switch {
		case len(meta.NestedFields) > 0:
			jsonExpr := traverseNestedStruct(meta, meta.NestedFields)
			if jsonExpr == "" {
				continue
			}
			*cols = append(*cols, fmt.Sprintf(`%s as "%s"`, jsonExpr, meta.JSONTag))

		// If there's no column tag, then column name will derive from JSON name
		case meta.ColumnTag == "":
			snake := CamelToSnake(meta.JSONTag)
			if snake == meta.JSONTag {
				*cols = append(*cols, meta.JSONTag)
			} else {
				*cols = append(*cols, fmt.Sprintf(`"%s" as "%s"`, snake, meta.JSONTag))
			}

		case meta.ColumnTag == "-":
			// Some columns are generated from jsonb_build_object,etc so it should be skipped here and use Select() instead
			continue

		// Just normal column name that use json as alias
		default:
			*cols = append(*cols, fmt.Sprintf(`%s as "%s"`, meta.ColumnTag, meta.JSONTag))
		}
	}

	return cols
}

// To generate nested field type with their respective sql aggregation (either jsonb_build_object or jsonb_agg).
// This function called in `buildColumnsFromMeta` when: field type is either struct (not time.Time) or slice of struct
// (these 2 conditions are marked by meta.NestedFields that has length above 0)
func traverseNestedStruct(parentField FieldMeta, nestedFieldMeta []FieldMeta) string {
	skippedJsonTag := []string{"", "-"}
	var exprs []string

	for i := range nestedFieldMeta {
		jsonTag := nestedFieldMeta[i].JSONTag
		columnTag := nestedFieldMeta[i].ColumnTag

		// Fast‑skip anything we don’t need to project.
		if ArrayIncludes(skippedJsonTag, jsonTag) || columnTag == "" {
			continue
		}

		exprs = append(exprs, fmt.Sprintf(`'%s', %s`, jsonTag, escapeQuoteColumns(columnTag)))
	}

	if len(exprs) == 0 {
		return ""
	}

	if parentField.IsSlice {
		return fmt.Sprintf("jsonb_agg(jsonb_build_object(%s))", strings.Join(exprs, ", "))
	}
	return fmt.Sprintf("jsonb_build_object(%s)", strings.Join(exprs, ", "))
}

func MakeMapJSONTags[T any]() map[string]string {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	// Get package path followed by struct name to give strict uniqueness to map's key
	typeName := typ.PkgPath() + "." + typ.Name()

	mu.Lock()
	defer mu.Unlock()
	// If columns not cached, fallback to process below
	fieldMeta, ok := fieldMetaCache[typeName]
	if !ok {
		// Build metadata if not cached (only once per type)
		meta := ExtractFromType(typ)
		fieldMetaCache[typeName] = &meta
		fieldMeta = &meta
	}

	return MakeMapJSONTagsFromType(*fieldMeta)
}

func MakeMapJSONTagsFromValue(dto any) map[string]string {
	typ := reflect.TypeOf(dto)
	if typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}

	// Get package path followed by struct name to give strict uniqueness to map's key
	typeName := typ.PkgPath() + "." + typ.Name()

	// If columns not cached, fallback to process below
	mu.Lock()
	defer mu.Unlock()
	fieldMeta, ok := fieldMetaCache[typeName]
	if !ok {
		// Build metadata if not cached (only once per type)
		meta := ExtractFromType(typ)
		fieldMetaCache[typeName] = &meta
		fieldMeta = &meta
	}

	return MakeMapJSONTagsFromType(*fieldMeta)
}

func MakeMapJSONTagsFromType(fieldMeta []FieldMeta) map[string]string {
	jsonMap := make(map[string]string)
	for i := 0; i < len(fieldMeta); i++ {
		field := fieldMeta[i]

		jsonTag := field.JSONTag
		columnTag := field.ColumnTag

		if jsonTag == "" || jsonTag == "-" || columnTag == "" {
			continue
		}

		jsonMap[jsonTag] = columnTag
	}
	return jsonMap
}
