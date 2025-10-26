package sql_query

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type UpdateRawSQL struct {
	Expr string
	Args []interface{}
}

// type string struct {
// 	Key       string
// 	Operator  SQLOperators // e.g., '=', '>', '<=', 'LIKE', 'IN', 'IS NULL'
// 	Value     interface{}  // could be a single value, slice, or nil
// 	IsRef     bool         // to determine whether WHERE is targeting literal value or reference
// 	ExtraArgs []interface{}
// }

type UpdateCases interface {
	// Case adds a WHEN clause to the current CASE expression defined by AddCase.
	// It appends the condition and its resulting value.
	//
	// Parameters:
	//   - conditions: a MultiFilterCondition containing AND and OR filters used for this WHEN clause.
	//   - value: the resulting value when the conditions match.
	//   - isRef: whether the value is a SQL column reference (true) or a literal that should be parameterized (false).
	//
	// Example:
	//   c.Case(MultiFilterCondition{
	//     And: map[string]SQLCondition{
	//       "status":       {Operator: SQLOperatorEqual, Value: "Emergency"},
	//       body.BloodType:   {SourceIsValue: true, Operator: SQLOperatorEqual, Value: "B"},
	//     },
	//   }, "Available", false)
	Case(conditions MultiFilterCondition, value interface{}, isRef bool)
	// Else sets the default ELSE value for the current CASE expression.
	// It is used when none of the WHEN conditions match.
	//
	// Parameters:
	//   - value: the fallback value if no WHEN clause matches.
	//   - isRef: whether the value is a SQL column reference (true) or a literal that should be parameterized (false).
	//
	// Example:
	//   c.Else("status", true) // fallback to the original "status" column value
	Else(value interface{}, isRef bool)
}

// To ensure SQLUpdateBuilder has its own initial methods
// e.g. UpdateBuilder(...).Update()...Rest
type SQLUpdateInitBuilder interface {
	// Update builds an UPDATE query for a single struct or map using reflection.
	//
	// It supports two value types:
	//   • Normal values (int, string, time.Time, etc.): converted into positional parameters ($1, $2, …).
	//   • Raw SQL expressions: allows embedding expressions with placeholders ("?")
	//     which will be replaced with proper PostgreSQL-style parameters ($1, $2, …).
	//
	// Example using struct:
	//
	//     type GroupQuota struct {
	//         Used int `column:"used"`
	//         Available sql_query.UpdateRawSQL `column:"available"`
	//     }
	//
	//     builder.Update(GroupQuota{
	//         Used: 5,
	//         Available: sql_query.UpdateRawSQL{
	//             Expr: "CASE WHEN available IS NULL THEN NULL ELSE available - ? END",
	//             Args: []any{5},
	//         },
	//     })
	//
	// Example using map:
	//
	//     builder.Update(map[string]any{
	//         "used": 5,
	//         "available": sql_query.UpdateRawSQL{
	//             Expr: "CASE WHEN available IS NULL THEN NULL ELSE available - ? END",
	//             Args: []any{5},
	//         },
	//     })
	//
	// By default, it will automatically add `updated_at = NOW()` if not explicitly provided.
	Update(values interface{}) SQLUpdateChainBuilder
	// UpdateEach updates multiple rows at once using VALUES() with a slice of structs.
	// Matches rows using the given rowIdentifier (e.g., "id").
	//
	// Example:
	//
	//	builder.UpdateEach([]User{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}, "id")
	//
	// → UPDATE users SET name = v.name, updated_at = NOW()
	//
	//	FROM (VALUES ($1,$2),($3,$4)) AS v(id,name,updated_at)
	//	WHERE users.id = v.id
	UpdateEach(values interface{}, rowIdentifier string) SQLUpdateChainBuilder

	// AddCase initializes a conditional CASE expression for the given column in an UPDATE statement.
	// It use completely different CASE expressions from previous AddCase and allows chaining multiple conditional branches using Case() and Else().
	//
	// Example:
	//   builder.AddCase("status", func(b UpdateCases) {
	//       b.Case(...)
	//       b.Else(...)
	//   })
	//
	// Parameters:
	//   - setColumn: the name of the column to be conditionally updated.
	//   - fn: a function that defines the CASE conditions using the provided UpdateCases interface.
	//
	// Returns:
	//   - SQLUpdateChainBuilder: the builder itself for method chaining.
	AddCase(setColumn string, fn func(b UpdateCases)) SQLUpdateChainBuilder

	// Increment is used to replace Update() for adding int value cause update cant handle that
	// Increment builds an UPDATE query that increases integer columns by a given value.
	// Automatically sets updated_at = NOW().
	//
	// Example:
	//
	//	builder.Increment(map[string]any{"count": 1})
	//
	// → UPDATE table SET "count" = "count" + $1, "updated_at" = NOW()
	Increment(values map[string]any) SQLUpdateChainBuilder

	// updateEachClausesGenerator looks at every struct in the slice and builds:
	//  1. The SET part of the query (e.g., "name = v.name"),
	//  2. The column names to use in the VALUES table,
	//  3. The placeholders ($1, $2, …) for each row.
	//
	// It's the core logic behind UpdateEach, letting you bulk-update rows
	// by matching them with a unique rowIdentifier (e.g. "id" or "name").
	//
	// This function loop through slice, returning generated clauses for everything in it
	// Expected result:
	// setClauses contains: [value = v.value, ...etc].
	// valueClauses contains: [value, ...etc].
	// valuePlaceholders contains: [($1, $2, $3, ...etc), ($4, $5, ...etc), ...etc]
	updateEachClausesGenerator(
		slice reflect.Value,
		rowIdentifier string,
	) ([]string, []string, []string)

	// extractUpdateFieldsStruct goes through a struct's fields and turns them into
	// SET clauses for an UPDATE query. It uses the "column" or "json" tag (or
	// the field name if tags are missing) to pick the column name.
	//
	// It also checks if an "updated_at" field exists, so the builder knows
	// whether to add "updated_at = NOW()" automatically.
	extractUpdateFieldsStruct(v reflect.Value) ([]string, bool)

	// extractUpdateFieldsMap takes a map[string]interface{} and returns a slice of
	// SET clauses for an UPDATE query. It uses the map keys as column names.
	// It also checks if an "updated_at" key exists, so the builder knows
	// whether to add "updated_at = NOW()" automatically.
	extractUpdateFieldsMap(v map[string]any) ([]string, bool)
}

// To ensure method .Update() has its own chaining methods
// e.g. .Update(...).From(...).Build()
type SQLUpdateChainBuilder interface {
	// AddCase initializes a conditional CASE expression for the given column in an UPDATE statement.
	// It clears any existing CASE expressions and allows chaining multiple conditional branches using Case() and Else().
	//
	// Example:
	//   builder.AddCase("status", func(b UpdateCases) {
	//       b.Case(...)
	//       b.Else(...)
	//   })
	//
	// Parameters:
	//   - setColumn: the name of the column to be conditionally updated.
	//   - fn: a function that defines the CASE conditions using the provided UpdateCases interface.
	//
	// Returns:
	//   - SQLUpdateChainBuilder: the builder itself for method chaining.
	AddCase(setColumn string, fn func(b UpdateCases)) SQLUpdateChainBuilder

	// By default, the Update builder includes all fields, even if their values are zero or nil.
	// Calling this method tells the builder to skip fields with zero values when generating the SET clause.
	//
	// Note: This option only affects single-row Update operations.
	// It has no effect on bulk UPDATE, because all rows in those operations must have the same set of columns.
	ExcludeEmpty() SQLUpdateChainBuilder
	// Update implements SQLUpdateChainBuilder. (Only able to be called once, will override previous call).
	// Conflict adds an ON CONFLICT clause with the specified constraint and action.
	//
	// Example:
	//
	//	builder.Conflict("(id)", "NOTHING")
	//
	// → ON CONFLICT (id) DO NOTHING
	Conflict(constraint, do string) SQLUpdateChainBuilder
	// Where implements SQLUpdateChainBuilder. (Accumulates previous value if called again)
	Where(filters map[string]SQLCondition) SQLUpdateChainBuilder
	// WhereOr implements SQLUpdateChainBuilder. (Accumulates previous value if called again)
	WhereOr(filters ...map[string]SQLCondition) SQLUpdateChainBuilder

	// Join adds an INNER JOIN clause with the specified ON condition.
	//
	// Example:
	//
	//	builder.Join("roles r", "r.id = u.role_id")
	Join(table string, onCondition string, additionalConditions ...map[string]SQLCondition) SQLUpdateChainBuilder
	// LeftJoin adds a LEFT JOIN clause with the specified ON condition.
	//
	// Example:
	//
	//	builder.LeftJoin("roles r", "r.id = u.role_id")
	LeftJoin(table string, onCondition string, additionalConditions ...map[string]SQLCondition) SQLUpdateChainBuilder

	// WithCTEBuilder adds a Common Table Expression (CTE) to the query.
	// It adjusts argument placeholders to avoid conflicts.
	// This function just add the defined CTE to the top of query.
	// You need to JOIN/LEFT JOIN the CTE builder to let main expression know that it should use CTE.
	//
	// Example:
	//
	//	builder.WithCTEBuilder("recent_orders", cte.(*sql_query.SelectBuilder).SQLEloquentQuery)
	//
	// Generates:
	//
	//	WITH recent_orders AS (SELECT id, user_id FROM orders) ...
	WithCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLUpdateChainBuilder

	// WithRecursiveCTEBuilder adds a Common Table Expression (CTE) to the query.
	// It adjusts argument placeholders to avoid conflicts.
	// This function just add the defined CTE to the top of query.
	// You need to JOIN/LEFT JOIN the CTE builder to let main expression know that it should use CTE.
	//
	// Example:
	//
	//	cte := NewSQLSelectBuilder[Order]("orders").Select("id", "user_id")
	//	builder.WithRecursiveCTEBuilder("recent_orders", cte.(*sql_query.SelectBuilder).SQLEloquentQuery)
	//
	// Generates:
	//
	//	WITH RECURSIVE recent_orders AS (SELECT id, user_id FROM orders) ...
	WithRecursiveCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLUpdateChainBuilder

	// Return implements SQLUpdateChainBuilder. (Only able to be called once, overrides previous values if re-called).
	// Return sets the columns to return after the update.
	// Defaults to RETURNING id if no column is provided.
	Return(columns ...string) SQLUpdateChainBuilder

	// From implements SQLUpdateChainBuilder. (Overrides previous value if called again)
	// From adds a FROM clause to the UPDATE query, allowing joins with other tables.
	//
	// Example:
	//
	//	builder.From([]string{"users u", "roles r"})
	//
	// → FROM users u, roles r
	From(tables []string) SQLUpdateChainBuilder

	// buildUpdateQuery constructs the final UPDATE query string and its arguments.
	// Ensures that CustomQuery is set and that a WHERE clause exists for safety.
	Build() (string, []interface{}, error)
}

type UpdateBuilder struct {
	*SQLEloquentQuery
}

func (s *UpdateBuilder) Where(filters map[string]SQLCondition) SQLUpdateChainBuilder {
	s.SQLEloquentQuery.sharedWhereAndQuery(filters)
	return s
}

func (s *UpdateBuilder) WhereOr(filters ...map[string]SQLCondition) SQLUpdateChainBuilder {
	s.SQLEloquentQuery.sharedWhereOrQuery(filters...)
	return s
}

func (s *UpdateBuilder) Join(
	table string,
	onCondition string,
	additionalConditions ...map[string]SQLCondition,
) SQLUpdateChainBuilder {
	if table == "" {
		return s
	}

	var filterSb strings.Builder
	if len(additionalConditions) > 0 {
		var filters []string
		s.sharedWhereAndQuery(additionalConditions[0], &filters)

		for i := range filters {
			filterSb.WriteString(" AND ")
			filterSb.WriteString(filters[i])
		}
	}

	s.OtherTables = append(s.OtherTables, fmt.Sprintf("JOIN %s ON %s%s", table, onCondition, filterSb.String()))
	return s
}

func (s *UpdateBuilder) LeftJoin(
	table string,
	mainCondition string,
	additionalConditions ...map[string]SQLCondition,
) SQLUpdateChainBuilder {
	if table == "" {
		return s
	}

	var filterSb strings.Builder
	if len(additionalConditions) > 0 {
		var filters []string
		s.sharedWhereAndQuery(additionalConditions[0], &filters)

		for i := range filters {
			filterSb.WriteString(" AND ")
			filterSb.WriteString(filters[i])
		}
	}

	s.OtherTables = append(s.OtherTables, fmt.Sprintf("LEFT JOIN %s ON %s%s", table, mainCondition, filterSb.String()))
	return s
}

func (s *UpdateBuilder) WithCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLUpdateChainBuilder {
	cteQuery, cteArgs, err := cteBuilder.Build()
	if err != nil {
		s.LastError = err
		return s
	}

	// Calculate the current argument offset
	offset := len(s.Args)

	// Shift the placeholders in the CTE query
	shiftedCTEQuery := shiftSQLPlaceholders(cteQuery, offset)

	s.WithClauses = append(s.WithClauses, fmt.Sprintf("%s AS (%s)", cteName, shiftedCTEQuery))
	s.Args = append(s.Args, cteArgs...)

	return s
}

func (s *UpdateBuilder) WithRecursiveCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLUpdateChainBuilder {
	cteQuery, cteArgs, err := cteBuilder.Build()
	if err != nil {
		s.LastError = err
		return s
	}

	// Calculate the current argument offset
	offset := len(s.Args)

	// Shift the placeholders in the CTE query
	shiftedCTEQuery := shiftSQLPlaceholders(cteQuery, offset)

	s.WithClauses = append(s.WithClauses, fmt.Sprintf("%s AS (%s)", cteName, shiftedCTEQuery))
	s.Args = append(s.Args, cteArgs...)

	s.useWithRecursive = true

	return s
}

func (s *UpdateBuilder) ExcludeEmpty() SQLUpdateChainBuilder {
	s.excludeEmptyValue = true
	return s
}

func (s *UpdateBuilder) Return(column ...string) SQLUpdateChainBuilder {
	if len(column) > 0 {
		s.Columns = column
	} else {
		s.Columns = []string{"id"}
	}
	return s
}

func (s *UpdateBuilder) Conflict(constraint, do string) SQLUpdateChainBuilder {
	s.ConflictClause = fmt.Sprintf(" ON CONFLICT %s DO %s", constraint, do)
	return s
}

func (s *UpdateBuilder) AddCase(setColumn string, fn func(b UpdateCases)) SQLUpdateChainBuilder {
	s.UpdateCaseClauses = make(map[string][]UpdateCaseParam, 0)
	s.UpdateCaseClauses[setColumn] = []UpdateCaseParam{}
	s.currentUpdateCaseKey = setColumn
	fn(s)
	return s
}

func (s *UpdateBuilder) Case(conditions MultiFilterCondition, value interface{}, isRef bool) {
	filters := []string{}

	if len(conditions.And) > 0 {
		andFilters := []string{}
		s.sharedWhereAndQuery(conditions.And, &andFilters)
		filters = append(filters, andFilters...)
	}

	if len(conditions.Or) > 0 {
		orFilters := []string{}
		s.whereOrDestination(conditions.Or, &orFilters)
		filters = append(filters, orFilters...)
	}

	var valueSb strings.Builder
	if isRef {
		valueSb.WriteString(value.(string))
	} else {
		valueSb.WriteByte('$')
		s.Args = append(s.Args, value)
		valueSb.WriteString(strconv.Itoa(len(s.Args)))
	}

	s.UpdateCaseClauses[s.currentUpdateCaseKey] = append(
		s.UpdateCaseClauses[s.currentUpdateCaseKey],
		UpdateCaseParam{conditions: filters, value: valueSb.String()},
	)
}

func (s *UpdateBuilder) Else(value interface{}, isRef bool) {
	var valueSb strings.Builder
	if isRef {
		valueSb.WriteString(value.(string))
	} else {
		valueSb.WriteByte('$')
		s.Args = append(s.Args, value)
		valueSb.WriteString(strconv.Itoa(len(s.Args)))
	}

	s.UpdateCaseClauses[s.currentUpdateCaseKey] = append(
		s.UpdateCaseClauses[s.currentUpdateCaseKey],
		UpdateCaseParam{conditions: []string{}, value: valueSb.String(), isElse: true},
	)
}

func (s *UpdateBuilder) Update(values interface{}) SQLUpdateChainBuilder {
	v := reflect.ValueOf(values)

	setClauses := []string{}
	hasUpdatedAt := false
	if v.Kind() == reflect.Struct {
		setClauses, hasUpdatedAt = s.extractUpdateFieldsStruct(v)
	} else if v.Kind() == reflect.Map {
		setClauses, hasUpdatedAt = s.extractUpdateFieldsMap(values.(map[string]any))
	} else {
		s.LastError = fmt.Errorf("invalid update values: expected struct or map, got %T", values)
		return s
	}

	if !hasUpdatedAt {
		setClauses = append(setClauses, `"updated_at" = NOW()`)
	}

	s.CustomQuery = fmt.Sprintf(`UPDATE %s SET %s`, s.Table, strings.Join(setClauses, ", "))

	return s
}

func (s *UpdateBuilder) UpdateEach(values interface{}, rowIdentifier string) SQLUpdateChainBuilder {
	v := reflect.ValueOf(values)

	// Slice data, plus length checking
	if v.Kind() != reflect.Slice || v.Len() == 0 {
		s.LastError = errors.New("update many values must be non-empty slice of struct")
		return s
	}

	// Check type of slice, should be slice of struct (only check the first index)
	firstElem := v.Index(0)
	if firstElem.Kind() != reflect.Struct {
		s.LastError = errors.New("update slice must contain structs")
		return s
	}

	// Main condition to make sure each row assigned by their respective value by matching row identifier
	// Example input key is "id"/"name", value is given in slice data with column tag same as input key (rowIdentifier), operator is equal
	// Assume v is aliased VALUES, example output id = v.id or could be name = v.name if this column is unique
	s.Filters = append(
		s.Filters,
		fmt.Sprintf(`%s."%s" = v."%s"`, s.Table, rowIdentifier, rowIdentifier),
	)

	// Generate clauses for everything in slice, then stored in main array
	var setClauses, valueClauses, placeholders []string
	mappedSet, mappedValue, mappedPlaceholders := s.updateEachClausesGenerator(v, rowIdentifier)
	setClauses = append(setClauses, mappedSet...)
	valueClauses = append(valueClauses, mappedValue...)
	placeholders = append(placeholders, mappedPlaceholders...)

	// Build everything into one query dedicated for update many
	s.CustomQuery = fmt.Sprintf(
		`UPDATE %s SET %s FROM (VALUES %s) as v(%s)`,
		s.Table,
		strings.Join(setClauses, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(valueClauses, ","),
	)

	return s
}

func (s *UpdateBuilder) Increment(
	values map[string]any,
) SQLUpdateChainBuilder {
	var setClauses []string

	for key, val := range values {
		snake := CamelToSnake(key)

		setClauses = append(
			setClauses,
			fmt.Sprintf(`"%s" = "%s" + $%d`, snake, snake, len(s.Args)+1),
		)
		s.Args = append(s.Args, val)
	}

	setClauses = append(setClauses, `"updated_at" = NOW()`)

	s.CustomQuery = fmt.Sprintf(`UPDATE %s SET %s`, s.Table, strings.Join(setClauses, ", "))
	return s
}

func (s *UpdateBuilder) From(tables []string) SQLUpdateChainBuilder {
	if len(tables) < 1 {
		return s
	}

	var otherTables []string
	otherTables = append(otherTables, fmt.Sprintf("FROM %s", strings.Join(tables, ", ")))
	s.OtherTables = otherTables
	return s
}

// NewSQLUpdateBuilder creates a new UpdateBuilder for the given table.
//
// Example:
//
//	builder := NewSQLUpdateBuilder("users","u")
//	builder.Update(user).Return("id").Build()
func NewSQLUpdateBuilder(tableName string, alias ...string) SQLUpdateInitBuilder {
	if len(alias) > 0 {
		tableName = fmt.Sprintf("%s %s", tableName, strings.TrimSpace(alias[0]))
	}

	return &UpdateBuilder{
		&SQLEloquentQuery{
			Table:       tableName,
			Filters:     []string{},
			OtherTables: []string{},
			Columns:     []string{},
			CustomQuery: "UPDATE",
			Args:        nil,
			Mode:        "update",
		},
	}
}

func (s *SQLEloquentQuery) buildUpdateQuery() (string, []interface{}, error) {
	if s.LastError != nil {
		return "", nil, errors.New(s.LastError.Error())
	}

	if s.CustomQuery == "" {
		return "", nil, errors.New("invalid update query: CustomQuery not set")
	}

	var initSb strings.Builder
	var withSb strings.Builder
	var whereSb strings.Builder
	var joinSb strings.Builder
	var fromSb strings.Builder
	var returningSb strings.Builder

	initSb.Grow(256)      // preallocate ~256 bytes
	withSb.Grow(256)      // preallocate ~256 bytes
	whereSb.Grow(256)     // preallocate ~256 bytes
	joinSb.Grow(256)      // preallocate ~256 bytes
	fromSb.Grow(256)      // preallocate ~256 bytes
	returningSb.Grow(256) // preallocate ~256 bytes

	initSb.WriteByte('\n')
	if len(s.UpdateCaseClauses) > 0 {
		initSb.WriteString(buildUpdateCase(s.UpdateCaseClauses, s.Table))
	} else {
		initSb.WriteString(s.CustomQuery)
	}

	// WITH
	if len(s.WithClauses) > 0 {
		withSb.WriteByte('\n')

		// if one of WITH are recursive, add 'RECURSIVE' string
		if s.useWithRecursive {
			withSb.WriteString("WITH RECURSIVE ")
		} else {
			withSb.WriteString("WITH ")
		}

		for i, c := range s.WithClauses {
			if i > 0 {
				withSb.WriteString(", ")
			}
			withSb.WriteString(c)
		}
	}

	if len(s.OtherTables) > 0 {
		fromSb.WriteByte('\n')
		fromSb.WriteString(strings.Join(s.OtherTables, " "))
	}

	if len(s.Filters) < 1 {
		return "", nil, errors.New("unsafe query: DELETE/UPDATE must have WHERE clause")
	}

	// WHERE
	whereSb.WriteByte('\n')
	whereSb.WriteString("WHERE ")
	for i, f := range s.Filters {
		if i > 0 {
			whereSb.WriteString(" AND ")
		}
		whereSb.WriteString(f)
	}
	whereSb.WriteByte('\n')

	if len(s.Columns) > 0 {
		returningSb.WriteByte('\n')
		returningSb.WriteString(" RETURNING " + strings.Join(s.Columns, ","))
	} else {
		returningSb.WriteByte('\n')
		returningSb.WriteString(" RETURNING id")
	}

	query := withSb.String() + initSb.String() + fromSb.String() + whereSb.String() + returningSb.String()

	return query, s.Args, nil
}

func (s *UpdateBuilder) updateEachClausesGenerator(
	slice reflect.Value,
	rowIdentifier string,
) ([]string, []string, []string) {
	var args []interface{}
	var setClauses, valueClauses, valuePlaceholders []string

	// Loop through slice
	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		t := v.Type()

		if i == 0 {
			// Setup column list once
			// Append set clauses and value clauses, also insert updated_at if not exists
			for j := 0; j < t.NumField(); j++ {
				field := t.Field(j)
				columnTag := field.Tag.Get("column")

				// If column tag is equal with rowIdentifier given by param, then it should not append into set clauses, only append to value clauses for WHERE condition
				if columnTag == rowIdentifier {
					valueClauses = append(valueClauses, fmt.Sprintf(`"%s"`, columnTag))
				} else {
					setClauses = append(setClauses, fmt.Sprintf(`"%s" = v."%s"`, columnTag, columnTag))
					valueClauses = append(valueClauses, fmt.Sprintf(`"%s"`, columnTag))
				}
			}

			if ok := ArrayIncludes(valueClauses, "updated_at"); !ok {
				setClauses = append(setClauses, `"updated_at" = NOW()`)
			}
		}

		var rowPlaceholders []string

		// Add fields
		for j := 0; j < t.NumField(); j++ {
			field := t.Field(j)
			transformTag := field.Tag.Get("transform")
			args = append(args, v.Field(j).Interface())
			rowPlaceholders = append(
				rowPlaceholders,
				fmt.Sprintf("$%d::%s", len(args), transformTag),
			)
		}

		// Append every placeholders to main array that will be returned
		valuePlaceholders = append(
			valuePlaceholders,
			fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", ")),
		)
	}

	s.Args = args
	return setClauses, valueClauses, valuePlaceholders
}

func (s *UpdateBuilder) extractUpdateFieldsStruct(v reflect.Value) ([]string, bool) {
	var setClauses []string
	var hasUpdatedAt bool

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)

		if val.IsZero() && s.excludeEmptyValue {
			continue
		}

		// Skip unexported or invalid fields
		if !val.IsValid() || !val.CanInterface() {
			continue
		}

		jsonTag := field.Tag.Get("json")

		specialTag := field.Tag.Get("special")

		// Handle ignored fields
		if jsonTag == "-" || strings.Contains(specialTag, "generated") {
			continue
		}

		// Use column tag > json tag > field name
		col := field.Tag.Get("column")
		if col == "" {
			if jsonTag != "" {
				col = CamelToSnake(jsonTag)
			} else {
				col = CamelToSnake(field.Name)
			}
		}
		// Slice from the first character all the way to the "." character
		if strings.Contains(col, ".") {
			col = col[strings.Index(col, ".")+1:]
		}

		// Recurse into nested structs (excluding time.Time)
		if val.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) &&
			field.Type != reflect.TypeOf(UpdateRawSQL{}) {
			childClauses, childHasUpdated := s.extractUpdateFieldsStruct(val)
			setClauses = append(setClauses, childClauses...)
			if childHasUpdated {
				hasUpdatedAt = true
			}
			continue
		}

		// If value is a map, recurse into it
		if val.Kind() == reflect.Map {
			childClauses, childHasUpdated := s.extractUpdateFieldsMap(
				val.Interface().(map[string]any),
			)
			setClauses = append(setClauses, childClauses...)
			if childHasUpdated {
				hasUpdatedAt = true
			}
			continue
		}

		fieldVal := val.Interface()

		switch v := fieldVal.(type) {
		case UpdateRawSQL:
			expr := v.Expr
			for _, arg := range v.Args {
				expr = strings.Replace(expr, "?", fmt.Sprintf("$%d", len(s.Args)+1), 1)
				s.Args = append(s.Args, arg)
			}
			setClauses = append(setClauses, fmt.Sprintf(`"%s" = %s`, col, expr))

		default:
			setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, col, len(s.Args)+1))
			s.Args = append(s.Args, fieldVal)
		}

		if col == "updated_at" {
			hasUpdatedAt = true
		}
	}

	return setClauses, hasUpdatedAt
}

func (s *UpdateBuilder) extractUpdateFieldsMap(v map[string]any) ([]string, bool) {
	setClauses := []string{}
	hasUpdatedAt := false

	for key, value := range v {
		isZeroValue := isZeroValue(value)
		if isZeroValue && s.excludeEmptyValue {
			continue
		}

		v := reflect.ValueOf(value)
		// if value is a struct, recurse into it
		if v.Kind() == reflect.Struct && v.Type() != reflect.TypeOf(time.Time{}) &&
			v.Type() != reflect.TypeOf(UpdateRawSQL{}) {
			childClauses, childHasUpdated := s.extractUpdateFieldsStruct(v)
			setClauses = append(setClauses, childClauses...)
			if childHasUpdated {
				hasUpdatedAt = true
			}
			continue
		}

		// if value is map, recurse into it
		if v.Kind() == reflect.Map {
			childClauses, childHasUpdated := s.extractUpdateFieldsMap(value.(map[string]any))
			setClauses = append(setClauses, childClauses...)
			if childHasUpdated {
				hasUpdatedAt = true
			}
			continue
		}

		// Use column name as key, or field name if not provided
		col := key
		if col == "updated_at" {
			hasUpdatedAt = true
		}

		switch v := value.(type) {
		case UpdateRawSQL:
			fmt.Println("v := value.(type) UpdateRawSQL", v)
			expr := v.Expr

			// replace ? with correct $n placeholders
			for _, arg := range v.Args {
				expr = strings.Replace(expr, "?", fmt.Sprintf("$%d", len(s.Args)+1), 1)
				s.Args = append(s.Args, arg)
			}
			setClauses = append(setClauses, fmt.Sprintf(`"%s" = %s`, col, expr))
		default:
			setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, col, len(s.Args)+1))
			s.Args = append(s.Args, v)
		}
	}

	return setClauses, hasUpdatedAt
}

// buildUpdateCase constructs a SQL UPDATE statement with CASE expressions.
//
// It takes a map of column names to slices of UpdateCaseParam, where each slice
// represents the WHEN/THEN/ELSE clauses for that column's CASE statement.
//
// Parameters:
//   - updateCaseClauses: A map where each key is a column name, and each value is a
//     slice of UpdateCaseParam defining the conditional logic for updating that column.
//   - tableName: The name of the table to update.
//
// Returns:
//   - A string containing the full SQL UPDATE query with conditional CASE logic.
//
// Example output:
//
//	UPDATE table_name
//	SET
//	  column1 = CASE
//	    WHEN condition1 AND condition2 THEN value1
//	    ELSE default_value
//	  END,
//	  updated_at = NOW()
//
// Notes:
//   - The function appends "updated_at = NOW()" to the final SET clause.
//   - It assumes all values and conditions are properly escaped/formatted.
func buildUpdateCase(updateCaseClauses map[string][]UpdateCaseParam, tableName string) string {
	var updateExpr string
	updateExpr = "UPDATE " + tableName + "\n"
	updateExpr += "SET\n"

	for key, each := range updateCaseClauses {
		updateExpr += key + " = CASE\n"
		fmt.Println("each", each)

		for _, param := range each {
			if param.isElse {
				updateExpr += "ELSE " + param.value
			} else {
				updateExpr += "WHEN " + strings.Join(param.conditions, " AND ") + " THEN " + param.value
			}
			updateExpr += "\n"
		}

		updateExpr += "END,\n"
	}

	updateExpr += "updated_at = NOW()"
	return updateExpr
}

func isZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}

	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
