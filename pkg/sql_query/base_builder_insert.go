package sql_query

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mystaline/clefinport-be/pkg/db"
)

// To ensure SQLInsertBuilder has its own initial methods
// e.g. InsertBuilder(...).Insert()...Rest
type SQLInsertInitBuilder interface {
	// Insert implements SQLInsertChainBuilder. (Only able to be called once)
	// Insert builds an INSERT query from either a single struct or a slice of structs.
	// - For a single struct → generates one row.
	// - For a slice → generates multiple rows.
	//
	// By default, it will RETURN the id::text column,
	// but you can pass custom returning columns:
	//
	//	Insert(user, "id", "name")
	//	Insert([]User{u1, u2})
	Insert(values interface{}, returningColumns ...string) SQLInsertChainBuilder

	// insertSingle handles the insert logic for a single struct.
	// It auto-generates a Snowflake ID and builds the VALUES list
	// based on struct field tags (using "column" or "json").
	cachedInsertSingle(v reflect.Value) SQLInsertChainBuilder
	// insertMany handles inserting multiple structs at once.
	// It generates a Snowflake ID for each row, creates placeholders
	// for every value, and builds the full VALUES (...), (...), ... list.
	//
	// Example Result:
	//
	//	INSERT INTO users (id, name, email)
	//	VALUES ($1, $2, $3), ($4, $5, $6)
	cachedInsertMany(slice reflect.Value) SQLInsertChainBuilder
	// preBuild is a helper that assembles the final
	// INSERT INTO ... (columns) VALUES ... statement
	// after the columns and value placeholders have been prepared.
	preBuild(columns, valuePlaceholders []string)
}

// To ensure method .Insert() has its own chaining methods
// e.g. .Insert(...).Build()
type SQLInsertChainBuilder interface {
	// By default, the Insert builder includes all fields, even if their values are zero or nil.
	// Calling this method tells the builder to skip fields with zero values when generating the SET clause.
	//
	// Note: This option only affects single-row Insert operations.
	// It has no effect on bulk INSERT, because all rows in those operations must have the same set of columns.
	ExcludeEmpty() SQLInsertChainBuilder
	// Insert implements SQLInsertChainBuilder. (Only able to be called once, will override previous call)
	// Conflict adds an ON CONFLICT clause to the insert statement.
	// Example:
	//
	//	.Conflict("(id)", "NOTHING")
	//	-> INSERT ... ON CONFLICT (id) DO NOTHING
	Conflict(constraint, do string) SQLInsertChainBuilder
	// buildInsertQuery finalizes the insert query into SQL string + args.
	// It prevents unsafe cases (like adding filters, joins, or pagination)
	// and appends RETURNING and ON CONFLICT if defined.
	Build() (string, []interface{}, error)
}

type InsertBuilder struct {
	*SQLEloquentQuery
}

func (s *InsertBuilder) ExcludeEmpty() SQLInsertChainBuilder {
	s.excludeEmptyValue = true
	return s
}

func (s *InsertBuilder) Conflict(constraint, do string) SQLInsertChainBuilder {
	s.ConflictClause = fmt.Sprintf(" ON CONFLICT %s DO %s", constraint, do)
	return s
}

func (s *InsertBuilder) Insert(
	values interface{},
	returningColumns ...string,
) SQLInsertChainBuilder {
	if len(returningColumns) > 0 {
		s.Columns = returningColumns
	} else {
		s.Columns = []string{"id"}
	}

	v := reflect.ValueOf(values)

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Struct {
		s.LastError = errors.New("insert values must be struct or slice of struct")
		return s
	}

	// Slice case
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			s.LastError = errors.New("cannot insert with empty slice")
			return s
		}

		firstElem := v.Index(0)
		if firstElem.Kind() != reflect.Struct {
			s.LastError = errors.New("insert slice must contain structs")
			return s
		}

		return s.cachedInsertMany(v)
	}

	// Single struct case
	return s.cachedInsertSingle(v)
}

// NewSQLInsertBuilder creates a new insert builder for a given table.
// Example:
//
//	builder := NewSQLInsertBuilder("users","u")
func NewSQLInsertBuilder(tableName string, alias ...string) SQLInsertInitBuilder {
	if len(alias) > 0 {
		tableName = fmt.Sprintf("%s %s", tableName, strings.TrimSpace(alias[0]))
	}

	return &InsertBuilder{
		&SQLEloquentQuery{
			Table:       tableName,
			Columns:     []string{},
			CustomQuery: "",
			Args:        nil,
			Mode:        "insert",
		},
	}
}

func (s *SQLEloquentQuery) buildInsertQuery() (string, []interface{}, error) {
	if s.LastError != nil {
		return "", nil, errors.New(s.LastError.Error())
	}

	if len(s.Filters) > 0 || s.UsePagination || len(s.OtherTables) > 0 {
		return "", nil, errors.New(
			"invalid insert query: cannot include filters, joins, or pagination",
		)
	}

	if s.ConflictClause != "" {
		s.CustomQuery += s.ConflictClause
	}

	if len(s.Columns) > 0 {
		s.CustomQuery += " RETURNING " + strings.Join(s.Columns, ",")
	}

	return s.CustomQuery, s.Args, nil
}

func (s *InsertBuilder) cachedInsertSingle(v reflect.Value) SQLInsertChainBuilder {
	t := v.Type()
	typeName := t.PkgPath() + "." + t.Name()

	// Get field meta from cache
	fieldMeta, ok := fieldMetaCache[typeName]
	if !ok {
		// Build metadata if not cached (only once per type)
		meta := ExtractFromType(t)
		fieldMetaCache[typeName] = &meta
		fieldMeta = &meta
	}

	// Get template for insert from cache
	cachedTemplate, ok := InsertCache[typeName]
	if !ok {
		// Build template if not cached (only once per type)
		var normalizedPlaceholders strings.Builder
		columns := []string{"id"}
		placeholders := []string{"($1"}
		FieldIndexes := [][]int{{0}}

		for _, meta := range *fieldMeta {
			// Skip generated column.
			if meta.IsGenerated {
				continue
			}

			if ArrayIncludes([]string{"_id", "id"}, meta.JSONTag) || meta.ColumnTag == "id" {
				continue
			}
			if ArrayIncludes([]string{"", "-"}, meta.JSONTag) &&
				ArrayIncludes([]string{"", "-"}, meta.ColumnTag) {
				continue
			}

			setTag := CamelToSnake(meta.JSONTag)
			if meta.ColumnTag != "" {
				if strings.Contains(meta.ColumnTag, ".") {
					// Slice from the first character all the way to the "." character
					setTag = meta.ColumnTag[strings.Index(meta.ColumnTag, ".")+1:]
				} else {
					setTag = meta.ColumnTag
				}
			}

			if ArrayIncludes([]string{"updated_at", "created_at"}, setTag) {
				continue
			}

			columns = append(columns, `"`+setTag+`"`)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(placeholders)+1))
			FieldIndexes = append(FieldIndexes, meta.FieldIndex)
		}

		columns = append(columns, "updated_at", "created_at")
		placeholders = append(placeholders, "NOW()", "NOW()")

		for idx, each := range placeholders {
			if idx > 0 {
				normalizedPlaceholders.WriteByte(',')
			}
			normalizedPlaceholders.WriteString(each)
			if idx == len(placeholders)-1 {
				normalizedPlaceholders.WriteByte(')')
			}
		}
		cachedTemplate = &InsertTemplate{
			InsertColumn:         columns,
			fieldMeta:            fieldMeta,
			singleRowPlaceholder: normalizedPlaceholders.String(),
			FieldIndexes:         FieldIndexes,
		}
		InsertCache[typeName] = cachedTemplate
	}

	args := make([]interface{}, 0, len(cachedTemplate.FieldIndexes)+1)
	db.InitSnowflake()

	for i, idx := range cachedTemplate.FieldIndexes {
		normalizedFieldMeta := *fieldMeta
		if len(args) == 0 && normalizedFieldMeta[i].Name != "ID" {
			args = append(args, db.Node.Generate().Int64()) // id in first position
			continue
		}

		val := v.FieldByIndex(idx)
		if len(args) != 0 && val.IsZero() && s.excludeEmptyValue {
			continue
		}

		if len(args) == 0 && val.IsZero() {
			args = append(args, db.Node.Generate().Int64()) // id in first position
		} else {
			args = append(args, val.Interface())
		}
	}

	s.Args = args
	s.preBuild(cachedTemplate.InsertColumn, []string{cachedTemplate.singleRowPlaceholder})

	return s
}

func (s *InsertBuilder) insertSingle(v reflect.Value) SQLInsertChainBuilder {
	var args []interface{}
	var columns []string
	var placeholders []string

	db.InitSnowflake() // prevents error in unit test (correct this if better method found) (Maybe mock snowflake idk)
	id := db.Node.Generate().Int64()
	columns = append(columns, "id")
	placeholders = append(placeholders, "$1")
	args = append(args, id)

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)

		if val.IsZero() && s.excludeEmptyValue {
			continue
		}

		jsonTag := field.Tag.Get("json")
		columnTag := field.Tag.Get("column")

		if ArrayIncludes([]string{"_id", "id"}, jsonTag) || columnTag == "id" {
			continue
		}

		if ArrayIncludes([]string{"", "-"}, jsonTag) &&
			ArrayIncludes([]string{"", "-"}, columnTag) {
			continue
		}

		setTag := CamelToSnake(jsonTag)
		if columnTag != "" {
			setTag = columnTag
		}

		columns = append(columns, `"`+setTag+`"`)

		if ArrayIncludes([]string{"updated_at", "created_at"}, setTag) {
			placeholders = append(placeholders, "NOW()")
			continue
		}

		args = append(args, val.Interface())
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}

	s.Args = args
	s.preBuild(columns, []string{fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))})

	return s
}

func (s *InsertBuilder) cachedInsertMany(slice reflect.Value) SQLInsertChainBuilder {
	db.InitSnowflake()
	firstElem := slice.Index(0)
	t := firstElem.Type()
	typeName := t.PkgPath() + ".[]" + t.Name()

	// Get field meta from cache
	fieldMeta, ok := fieldMetaCache[typeName]
	if !ok {
		// Build metadata if not cached (only once per type)
		meta := ExtractFromType(t)
		fieldMetaCache[typeName] = &meta
		fieldMeta = &meta
	}

	// Get template for insert from cache
	cachedTemplate, ok := InsertCache[typeName]
	if !ok {
		// Build template if not cached (only once per type)
		columns := []string{"id"}
		FieldIndexes := make([][]int, 0, len(*fieldMeta))
		basePlaceholders := make([]string, 0, len(*fieldMeta)+1)

		// ID in first position
		basePlaceholders = append(basePlaceholders, "$1")
		FieldIndexes = append(FieldIndexes, []int{0})

		for _, meta := range *fieldMeta {
			// Skip generated column.
			if meta.IsGenerated {
				continue
			}

			if ArrayIncludes([]string{"_id", "id"}, meta.JSONTag) || meta.ColumnTag == "id" {
				continue
			}
			if ArrayIncludes([]string{"", "-"}, meta.JSONTag) &&
				ArrayIncludes([]string{"", "-"}, meta.ColumnTag) {
				continue
			}

			setTag := CamelToSnake(meta.JSONTag)
			if meta.ColumnTag != "" {
				if strings.Contains(meta.ColumnTag, ".") {
					// Slice from the first character all the way to the "." character
					setTag = meta.ColumnTag[strings.Index(meta.ColumnTag, ".")+1:]
				} else {
					setTag = meta.ColumnTag
				}
			}
			if ArrayIncludes([]string{"updated_at", "created_at"}, setTag) {
				continue
			}

			columns = append(columns, `"`+setTag+`"`)
			FieldIndexes = append(FieldIndexes, meta.FieldIndex)
			basePlaceholders = append(basePlaceholders, "$X")
		}

		columns = append(columns, "updated_at", "created_at")
		basePlaceholders = append(basePlaceholders, "NOW()", "NOW()")

		cachedTemplate = &InsertTemplate{
			InsertColumn:     columns,
			basePlaceholders: basePlaceholders,
			FieldIndexes:     FieldIndexes,
		}
		InsertCache[typeName] = cachedTemplate
	}

	numRows := slice.Len()
	numCols := len(cachedTemplate.basePlaceholders)

	args := make([]interface{}, numRows*numCols)
	var allPlaceholders strings.Builder
	allPlaceholders.Grow(numRows * (numCols * 4)) // pre-allocate

	argsPos := 0
	for i := 0; i < numRows; i++ {
		values := slice.Index(i)
		if i > 0 {
			allPlaceholders.WriteByte(',')
		}
		allPlaceholders.WriteByte('(')

		for j, base := range cachedTemplate.basePlaceholders {
			if j > 0 {
				allPlaceholders.WriteByte(',')
			}

			if base == "NOW()" {
				allPlaceholders.WriteString("NOW()")
				continue
			}

			normalizedFieldMeta := *fieldMeta
			if base == "$1" && normalizedFieldMeta[j].Name != "ID" {
				args[argsPos] = db.Node.Generate().Int64() // id in first position
			} else {
				val := values.FieldByIndex(cachedTemplate.FieldIndexes[j])
				if base == "$1" && val.IsZero() {
					args[argsPos] = db.Node.Generate().Int64()
				} else {
					args[argsPos] = val.Interface()
				}
			}

			argsPos++
			allPlaceholders.WriteByte('$')
			allPlaceholders.WriteString(strconv.Itoa(argsPos))
		}

		allPlaceholders.WriteByte(')')
	}

	s.Args = args[:argsPos] // Trim args to the actual used size because there might excluded columns
	s.preBuild(cachedTemplate.InsertColumn, []string{allPlaceholders.String()})

	return s
}

func (s *InsertBuilder) insertMany(slice reflect.Value) SQLInsertChainBuilder {
	var args []interface{}
	var columns []string
	var valuePlaceholders []string

	db.InitSnowflake() // prevents error in unit test (correct this if better method found) (Maybe mock snowflake idk)

	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		t := v.Type()

		if i == 0 {
			// Setup column list once
			columns = append(columns, "id")
			for j := 0; j < t.NumField(); j++ {
				field := t.Field(j)
				jsonTag := field.Tag.Get("json")
				columnTag := field.Tag.Get("column")

				if ArrayIncludes([]string{"_id", "id"}, jsonTag) || columnTag == "id" {
					continue
				}

				if ArrayIncludes([]string{"", "-"}, jsonTag) &&
					ArrayIncludes([]string{"", "-"}, columnTag) {
					continue
				}

				setTag := CamelToSnake(jsonTag)
				if columnTag != "" {
					setTag = columnTag
				}
				columns = append(columns, `"`+setTag+`"`)
			}
		}

		var rowPlaceholders []string

		// Add ID
		id := db.Node.Generate().Int64()
		args = append(args, id)
		startIndex := len(args)
		rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", startIndex))

		// Add fields
		for j := 0; j < t.NumField(); j++ {
			field := t.Field(j)
			jsonTag := field.Tag.Get("json")
			columnTag := field.Tag.Get("column")
			setTag := CamelToSnake(jsonTag)
			if columnTag != "" {
				setTag = columnTag
			}

			if ArrayIncludes([]string{"_id", "id"}, jsonTag) || columnTag == "id" {
				continue
			}

			if ArrayIncludes([]string{"", "-"}, jsonTag) &&
				ArrayIncludes([]string{"", "-"}, columnTag) {
				continue
			}

			if ArrayIncludes([]string{"updated_at", "created_at"}, setTag) {
				rowPlaceholders = append(rowPlaceholders, "NOW()")
				continue
			}

			args = append(args, v.Field(j).Interface())
			rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", len(args)))
		}

		valuePlaceholders = append(
			valuePlaceholders,
			fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", ")),
		)
	}

	s.Args = args
	s.preBuild(columns, valuePlaceholders)

	return s
}

func BuildInsertTemplate(t reflect.Type) *InsertTemplate {
	meta := ExtractFromType(t)

	columns := []string{}
	fieldIndexes := [][]int{}
	useID := []bool{}
	useNow := []bool{}

	useID = append(useID, true)
	useNow = append(useNow, false)
	fieldIndexes = append(fieldIndexes, nil)

	for _, m := range meta {
		if ArrayIncludes([]string{"_id", "id"}, m.JSONTag) || m.ColumnTag == "id" {
			continue
		}
		if ArrayIncludes([]string{"", "-"}, m.JSONTag) &&
			ArrayIncludes([]string{"", "-"}, m.ColumnTag) {
			continue
		}

		setTag := CamelToSnake(m.JSONTag)
		if m.ColumnTag != "" {
			setTag = m.ColumnTag
		}

		columns = append(columns, setTag)

		if setTag == "created_at" || setTag == "updated_at" {
			useID = append(useID, false)
			useNow = append(useNow, true)
			fieldIndexes = append(fieldIndexes, nil)
			continue
		}

		useID = append(useID, false)
		useNow = append(useNow, false)
		fieldIndexes = append(fieldIndexes, m.FieldIndex)
	}

	return &InsertTemplate{
		InsertColumn: columns,
		FieldIndexes: fieldIndexes,
		UseID:        useID,
		UseNow:       useNow,
	}
}

func (s *InsertBuilder) preBuild(columns, valuePlaceholders []string) {
	var sb strings.Builder
	sb.Grow(256) // preallocate ~256 bytes

	sb.WriteByte('\n')
	sb.WriteString("INSERT INTO ")
	sb.WriteString(s.Table)
	sb.WriteByte('\n')
	for i, column := range columns {
		if i == 0 {
			sb.WriteByte('(')
		}

		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(column)

		if i == len(columns)-1 {
			sb.WriteByte(')')
		}
	}

	sb.WriteByte('\n')
	sb.WriteString("VALUES ")
	for i, placeholder := range valuePlaceholders {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(placeholder)
	}
	sb.WriteByte('\n')

	s.CustomQuery = sb.String()
	// fmt.Println("s.Args", s.Args)
	// fmt.Println("s.CustomQuery", s.CustomQuery)
}
