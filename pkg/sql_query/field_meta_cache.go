package sql_query

import (
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type FieldMeta struct {
	Name         string
	Type         reflect.Type
	JSONTag      string
	ColumnTag    string
	SQLExpr      string
	FieldIndex   []int
	NestedFields []FieldMeta

	IsStruct    bool
	IsSlice     bool
	IsTime      bool
	IsGenerated bool
}

type InsertTemplate struct {
	InsertColumn         []string
	FieldIndexes         [][]int
	fieldMeta            *[]FieldMeta
	basePlaceholders     []string
	singleRowPlaceholder string
	UseID                []bool
	UseNow               []bool
}

type CacheKey struct {
	Typ    reflect.Type
	Colsig string
}

var (
	fieldMetaCache = make(map[string]*[]FieldMeta)
	columnsCache   = make(map[string]*[]string)

	InsertCache = make(map[string]*InsertTemplate)

	FieldMapCache sync.Map
)

func normalizeType(t reflect.Type) (normalizedType reflect.Type, isSlice bool) {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		if t.Kind() == reflect.Slice {
			isSlice = true
		}
		t = t.Elem()
	}
	normalizedType = t
	return normalizedType, isSlice
}

func ExtractFromType(typ reflect.Type) []FieldMeta {
	var fields []FieldMeta
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		if f.PkgPath != "" {
			continue
		}

		specialTag := f.Tag.Get("special")

		var isSlice bool
		fType := f.Type
		fType, isSlice = normalizeType(fType)
		timeType := reflect.TypeOf(time.Time{})

		meta := FieldMeta{
			Name:        f.Name,
			Type:        fType,
			JSONTag:     strings.TrimSuffix(f.Tag.Get("json"), ",omitempty"),
			ColumnTag:   f.Tag.Get("column"),
			FieldIndex:  f.Index,
			IsStruct:    fType.Kind() == reflect.Struct && fType != timeType,
			IsSlice:     isSlice,
			IsTime:      fType == timeType,
			IsGenerated: strings.Contains(specialTag, "generated"),
		}

		// No longer need to check for .Elem() since variable `meta` already normalized to single struct
		if meta.IsStruct || (meta.IsSlice && meta.Type.Kind() == reflect.Struct) {
			meta.NestedFields = ExtractFromType(meta.Type)
		}

		fields = append(fields, meta)
	}

	return fields
}

func GetFieldMap(elemType reflect.Type, fds []pgconn.FieldDescription) []int {
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	key := CacheKey{Typ: elemType, Colsig: columnsSignature(fds)}
	if cached, ok := FieldMapCache.Load(key); ok {
		return cached.([]int)
	}

	// Build lookup table for struct fields
	lookup := map[string]int{}
	for i := 0; i < elemType.NumField(); i++ {
		f := elemType.Field(i)
		if f.PkgPath != "" {
			continue // skip unexported fields
		}

		name := f.Tag.Get("column")
		if name == "" {
			name = f.Name
		}
		lookup[strings.ToLower(name)] = i
	}

	indices := make([]int, len(fds))
	for i, fd := range fds {
		if idx, ok := lookup[strings.ToLower(string(fd.Name))]; ok {
			indices[i] = idx
		} else {
			indices[i] = -1
		}
	}

	FieldMapCache.Store(key, indices)
	return indices
}

func columnsSignature(fds []pgconn.FieldDescription) string {
	parts := make([]string, len(fds))
	for i, fd := range fds {
		parts[i] = string(fd.Name)
	}
	return strings.Join(parts, ",")
}

func ClearCache() {
	fieldMetaCache = make(map[string]*[]FieldMeta)
	columnsCache = make(map[string]*[]string)
	InsertCache = make(map[string]*InsertTemplate)
}
