package sql_query

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type SQLSelectChainBuilder interface {
	// GetCurrentArgIndex returns the current number of arguments in the query.
	// Useful for calculating placeholder positions when building parameterized SQL.
	GetCurrentArgIndex() int
	// AddArgs appends one or more arguments to the query's argument list.
	// Can be used to manually manage parameter placeholders.
	AddArgs(arg ...interface{}) SQLSelectChainBuilder
	// StartPlaceholderFrom set starting point of next generated placeholders.
	// Useful for creating sub query from this builder for another main query.
	// Placeholders for this sub query will be started from this given value.
	StartPlaceholderFrom(index int) SQLSelectChainBuilder

	// Distinct implements SQLSelectChainBuilder.
	// Distinct defines one or more columns for the DISTINCT ON(...) statement.
	//
	// Example:
	//
	//	builder.Distinct("distinct", "u.id", "u.name")
	Distinct(alias string, columns ...string) SQLSelectChainBuilder

	// Select implements SQLSelectChainBuilder.
	// Select defines one or more columns for the SELECT statement.
	// If a column alias already exists, its expression will be overwritten with the new one.
	//
	// Example:
	//
	//	builder.Select("u.id AS user_id", "u.name")
	Select(columns ...string) SQLSelectChainBuilder

	// USE WITH CAUTION
	// Reset all previous appended selects
	// Useful for this part in GenerateCTEOption
	// cteBuilder := sourceBuilder.
	//		ClearSelects(). -- without this the previous selects inside the source builder will get included
	// 		Distinct(
	// 			fmt.Sprintf(`%s AS "value"`, labelValue),
	// 			labelValue,
	// 		).
	// 		Select(
	// 			fmt.Sprintf(`%s AS "key"`, labelKey),
	// 			fmt.Sprintf(`"%s"."id" AS "id"`, refTable),
	// 		)
	//
	// Example usage:
	//
	//	builder.ClearSelect()
	ClearSelects() SQLSelectChainBuilder

	// SelectCaseWhen adds a CASE WHEN expression as a column.
	//
	// Example:
	//
	//	builder.SelectCaseWhen("'Yes'", "'No'", "is_admin", "role = 'admin'")
	//
	// Generates:
	//
	//	CASE WHEN role = 'admin' THEN 'Yes' ELSE 'No' END AS is_admin
	SelectCaseWhen(thenExpr, elseExpr, alias string, whenClause string, whenArgs ...interface{}) SQLSelectChainBuilder
	// SelectBoolAnd adds a bool_or aggregate column with an alias.
	//
	// Example:
	//
	//	builder.SelectBoolAnd("is_active", "any_active")
	//
	// Generates:
	//
	//	bool_and(is_active) AS any_active
	SelectBoolAnd(expr, alias string, args ...interface{}) SQLSelectChainBuilder
	// SelectBoolOr adds a bool_or aggregate column with an alias.
	//
	// Example:
	//
	//	builder.SelectBoolOr("is_active", "any_active")
	//
	// Generates:
	//
	//	bool_or(is_active) AS any_active
	SelectBoolOr(expr, alias string, args ...interface{}) SQLSelectChainBuilder

	SelectArrayAggregation(alias string, source string, config ArrayAggConfig) SQLSelectChainBuilder

	// SelectJSONArrayElements selects elements from a Go slice of maps
	// and expands them as rows using jsonb_array_elements().
	//
	// Each element in the slice is marshaled to JSON and passed as a query argument.
	//
	// Example:
	//
	//	arr := []map[string]string{{"id": "1"}, {"id": "2"}}
	//	builder.SelectJSONArrayElements("items", arr)
	//
	// Generates:
	//
	//	jsonb_array_elements($1::jsonb) AS items
	SelectJSONArrayElements(alias string, arrayElements []map[string]string) SQLSelectChainBuilder
	// SelectJSONAggregate builds a JSON object or JSON array aggregation using jsonb_build_object
	// or jsonb_agg(jsonb_build_object(...)). It supports optional filtering and ordering.
	//
	// The `dto` parameter can be either:
	//   - a struct (fields are mapped from JSON tags), or
	//   - a map[string]string (keys = JSON keys, values = SQL expressions).
	//
	// If asArrayAggregation is true, the result is wrapped with jsonb_agg().
	//
	// Example (array aggregation):
	//
	//	builder.SelectJSONAggregate(
	//	    "order_items",
	//	    map[string]string{"id": "items.id", "name": "items.name"},
	//	    "items.is_active = TRUE",
	//	    true,
	//	    "items.created_at",
	//	)
	//
	// or with a struct:
	//
	//	builder.SelectJSONAggregate(
	//	    "order_items",
	//	    dto.OrderItem{},
	//	    "items.is_active = TRUE",
	//	    true,
	//	    "items.created_at",
	//	)
	//
	// Generates:
	//
	//	jsonb_agg(jsonb_build_object('id', items.id, 'name', items.name) ORDER BY items.created_at)
	//	  FILTER (WHERE items.is_active = TRUE) AS order_items
	SelectJSONAggregate(alias string, dto any, condition string, asArrayAggregation bool, orderByClauses ...string) SQLSelectChainBuilder
	// Read documentation for SelectJSONAggregate since the function is similar but with additional COALESCE
	// Generates:
	//
	//	COALESCE(jsonb_agg(DISTINCT jsonb_build_object('id', items.id, 'name', items.name) ORDER BY items.created_at)
	//	  FILTER (WHERE items.is_active = TRUE) AS order_items, ${coalesce})
	SelectJSONAggregateCoalesce(alias string, dto any, condition string, asArrayAggregation bool, coalesce string, orderByClauses ...string) SQLSelectChainBuilder
	// Read documentation for SelectJSONAggregate since the function is similar but with additional DISTINCT
	// Generates:
	//
	//	jsonb_agg(DISTINCT jsonb_build_object('id', items.id, 'name', items.name) ORDER BY items.created_at)
	//	  FILTER (WHERE items.is_active = TRUE) AS order_items
	SelectJSONAggregateDistinct(alias string, dto any, condition string, asArrayAggregation bool, orderByClauses ...string) SQLSelectChainBuilder
	// SelectJSONAggregateFunc builds a nested JSON object by executing a callback
	// that itself adds JSON aggregate fields. The resulting fields are combined
	// into a single jsonb_build_object aliased as `alias`.
	//
	// Example:
	//
	//	builder.SelectJSONAggregateFunc("hasSystemRole", func(b *sql_query.SelectBuilder) {
	//	    b.SelectJSONAggregate(
	//	        string(rt),
	//	        map[string]string{
	//	            "create": getSystemRolePermission("create"),
	//	            "view":   getSystemRolePermission("view"),
	//	            "update": getSystemRolePermission("update"),
	//	            "delete": getSystemRolePermission("delete"),
	//	        },
	//	        fmt.Sprintf("usr.role_attribute = $%d", len(builder.Args)+1),
	//	        false,
	//	    )
	//	})
	//
	// Generates:
	//
	//	jsonb_build_object(
	//	    'SomeRole', CASE WHEN usr.role_attribute = $1 THEN jsonb_build_object(
	//	        'create', <expr>, 'view', <expr>, 'update', <expr>, 'delete', <expr>
	//	    ) ELSE NULL END
	//	) AS hasSystemRole
	SelectJSONAggregateFunc(alias string, fn func(builder *SelectBuilder)) SQLSelectChainBuilder

	// Where implements SQLSelectChainBuilder. (Accumulates previous value if called again)
	Where(filters map[string]SQLCondition) SQLSelectChainBuilder
	// WhereOr implements SQLSelectChainBuilder. (Accumulates previous value if called again)
	WhereOr(filters ...map[string]SQLCondition) SQLSelectChainBuilder

	// Search implements SQLSelectChainBuilder and accumulates conditions if called multiple times.
	// Adds a case-insensitive ILIKE condition across multiple fields, combined with OR.
	//
	// To search inside array columns, add suffix ":array" to the column name.
	// This uses EXISTS + unnest for better performance & usabilitiy than array_to_string.
	//
	// Example:
	//
	//	builder.Search("john", []string{"status::text", "last_name", "tags:array"})
	//
	// Generates:
	//
	//	(first_name ILIKE $1 OR last_name ILIKE $1 OR EXISTS (SELECT 1 FROM unnest(tags) AS val WHERE val ILIKE $1))
	Search(keyword string, fields []string) SQLSelectChainBuilder
	// SetLimit sets a fixed LIMIT value for the query (overwrites any previous limit).
	//
	// Example:
	//
	//	builder.SetLimit(5)
	SetLimit(limit int) SQLSelectChainBuilder
	// Join adds an INNER JOIN clause with the specified ON condition.
	//
	// Example:
	//
	//	builder.Join("roles r", "r.id = u.role_id")
	Join(table string, onCondition string, additionalConditions ...map[string]SQLCondition) SQLSelectChainBuilder
	// LeftJoin adds a LEFT JOIN clause with the specified ON condition.
	//
	// Example:
	//
	//	builder.LeftJoin("roles r", "r.id = u.role_id")
	LeftJoin(table string, onCondition string, additionalConditions ...map[string]SQLCondition) SQLSelectChainBuilder

	// Example:
	//
	// LeftJoinLateralWithQuery("ta", categoryWithRecursiveBuilder().(*sql_query.SelectBuilder).SQLEloquentQuery, "TRUE").
	//
	// Output:
	//
	// 	LEFT JOIN LATERAL (
	//     `Sql queries here`
	//   ) `alias` ON `condition`
	LeftJoinLateralWithQuery(joinName string, joinQueryBuilder *SQLEloquentQuery, mainCondition string, additionalConditions ...map[string]SQLCondition) SQLSelectChainBuilder

	// Paginate implements SQLSelectChainBuilder. (Overrides previous value if called again).
	// Paginate applies LIMIT, OFFSET, and ORDER BY using a Pagination struct.
	// It supports single or multiple sorting rules.
	//
	// Example:
	//
	//	builder.Paginate(Pagination{Page: 2, Limit: 10, SortBy: "name", SortOrder: 1})
	Paginate(query Pagination) SQLSelectChainBuilder
	// OrderBy adds sorting rules to the query. Multiple calls accumulate sorting.
	//
	// Example:
	//
	//	builder.OrderBy([]string{"created_at"}, false) // DESC
	OrderBy(sortBy []string, asc bool) SQLSelectChainBuilder
	// GroupBy adds one or more columns to the GROUP BY clause.
	// Multiple calls accumulate columns.
	//
	// Example:
	//
	//	builder.GroupBy("department", "role")
	GroupBy(groupBy ...string) SQLSelectChainBuilder
	// Having implements SQLSelectChainBuilder. (Overrides previous value if called again).
	// Having adds a HAVING clause for grouped queries.
	// Overwrites any previous HAVING condition.
	//
	// Example:
	//
	//	builder.GroupBy("role").Having(map[string]SQLCondition{
	//	    "count(*)": {Op: ">", Value: 5},
	//	})
	Having(havingClauses map[string]SQLCondition) SQLSelectChainBuilder

	// WithCTEBuilder adds a Common Table Expression (CTE) to the query.
	// It adjusts argument placeholders to avoid conflicts.
	// This function just add the defined CTE to the top of query.
	// You need to JOIN/LEFT JOIN the CTE builder to let main expression know that it should use CTE.
	//
	// Example:
	//
	//	cte := NewSQLSelectBuilder[Order]("orders").Select("id", "user_id")
	//	builder.WithCTEBuilder("recent_orders", cte.(*sql_query.SelectBuilder).SQLEloquentQuery)
	//
	// Generates:
	//
	//	WITH recent_orders AS (SELECT id, user_id FROM orders) ...
	WithCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLSelectChainBuilder

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
	WithRecursiveCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLSelectChainBuilder

	// Add "UNION ALL" in between the queries
	// example:
	//  SELECT id
	// FROM categories
	// WHERE id = 0
	//
	// UNION ALL
	//
	// SELECT c.id
	// FROM categories c
	// INNER JOIN category_tree ct ON c.parent_id = ct.id
	UnionAll(cteBuilders ...*SQLEloquentQuery) SQLSelectChainBuilder

	// Build finalizes the SELECT query and returns the query string and arguments.
	// Returns an error if the query is invalid (e.g., HAVING without GROUP BY).
	Build() (string, []interface{}, error)
}

type SelectBuilder struct {
	*SQLEloquentQuery
}

func (s *SelectBuilder) Where(filters map[string]SQLCondition) SQLSelectChainBuilder {
	s.SQLEloquentQuery.sharedWhereAndQuery(filters)
	return s
}

func (s *SelectBuilder) WhereOr(filters ...map[string]SQLCondition) SQLSelectChainBuilder {
	s.SQLEloquentQuery.sharedWhereOrQuery(filters...)
	return s
}

func (s *SelectBuilder) GetCurrentArgIndex() int {
	return len(s.Args)
}

func (s *SelectBuilder) AddArgs(arg ...interface{}) SQLSelectChainBuilder {
	s.Args = append(s.Args, arg...)
	return s
}

func (s *SelectBuilder) StartPlaceholderFrom(index int) SQLSelectChainBuilder {
	if index < 1 {
		index = 1
	}

	if len(s.Args) > 0 {
		// Need to pad so that first existing arg becomes $index
		pad := index - 1
		if pad > 0 {
			temp := make([]interface{}, pad+len(s.Args))
			// fill padding with nils
			for i := 0; i < pad; i++ {
				temp[i] = nil
			}
			// copy old args after padding
			copy(temp[pad:], s.Args)
			s.Args = temp
		}
	} else {
		// If empty, just pad with nils up to index-1
		if index > 1 {
			temp := make([]interface{}, index-1)
			for i := 0; i < index-1; i++ {
				temp[i] = nil
			}
			s.Args = temp
		}
	}

	return s
}

func (s *SelectBuilder) Search(keyword string, fields []string) SQLSelectChainBuilder {
	if keyword != "" && len(fields) > 0 {
		var orClauses []string
		orArgs := []interface{}{}

		for _, field := range fields {
			isArrayColumn := strings.HasSuffix(field, ":array")

			if isArrayColumn {
				cleanColumn := strings.TrimSuffix(field, ":array")
				orClauses = append(orClauses, fmt.Sprintf("EXISTS (SELECT 1 FROM unnest(%s) as val WHERE val ILIKE $%d)", cleanColumn, len(s.Args)+len(orArgs)+1))
			} else {
				orClauses = append(orClauses, fmt.Sprintf("%s ILIKE $%d", field, len(s.Args)+len(orArgs)+1))
			}

			orArgs = append(orArgs, "%"+keyword+"%")
		}

		// Combine all clauses with OR inside parentheses
		orClauseGroup := fmt.Sprintf("(%s)", strings.Join(orClauses, " OR "))
		s.Filters = append(s.Filters, orClauseGroup)
		s.Args = append(s.Args, orArgs...)
	}

	return s
}

func (s *SelectBuilder) Distinct(alias string, columns ...string) SQLSelectChainBuilder {
	s.DistinctAlias = alias
	for _, newCol := range columns {
		newAlias := extractAlias(newCol)

		// Check if alias exists in current list
		replaced := false
		for i, existing := range s.DistinctBy {
			extracted := extractAlias(existing)
			if extracted != "" && extracted == newAlias {
				s.DistinctBy[i] = newCol // Overwrite
				replaced = true
				break
			}
		}

		if !replaced {
			s.DistinctBy = append(s.DistinctBy, newCol)
		}
	}

	return s
}

func (s *SelectBuilder) ClearSelects() SQLSelectChainBuilder {
	s.Columns = []string{}
	return s
}

func (s *SelectBuilder) Select(columns ...string) SQLSelectChainBuilder {
	for _, newCol := range columns {
		newAlias := extractAlias(newCol)

		// Check if alias exists in current list
		replaced := false
		for i, existing := range s.Columns {
			extracted := extractAlias(existing)
			if extracted != "" && extracted == newAlias {
				s.Columns[i] = newCol // Overwrite
				replaced = true
				break
			}
		}

		if !replaced {
			s.Columns = append(s.Columns, newCol)
		}
	}
	return s
}

func (s *SelectBuilder) SelectBoolAnd(expr, alias string, args ...interface{}) SQLSelectChainBuilder {
	boolAndColumn := fmt.Sprintf("bool_and(%s) AS \"%s\"", expr, alias)

	// Check if alias exists in current list
	replaced := false
	for i, existing := range s.Columns {
		extracted := extractAlias(existing)
		if extracted != "" && extracted == strings.ToLower(alias) {
			s.Columns[i] = boolAndColumn // Overwrite
			replaced = true
			break
		}
	}

	if !replaced {
		s.Columns = append(s.Columns, boolAndColumn)
	}

	s.Args = append(s.Args, args...)
	return s
}

func (s *SelectBuilder) SelectBoolOr(expr, alias string, args ...interface{}) SQLSelectChainBuilder {
	boolOrColumn := fmt.Sprintf("bool_or(%s) AS \"%s\"", expr, alias)

	// Check if alias exists in current list
	replaced := false
	for i, existing := range s.Columns {
		extracted := extractAlias(existing)
		if extracted != "" && extracted == strings.ToLower(alias) {
			s.Columns[i] = boolOrColumn // Overwrite
			replaced = true
			break
		}
	}

	if !replaced {
		s.Columns = append(s.Columns, boolOrColumn)
	}

	s.Args = append(s.Args, args...)
	return s
}

func (s *SelectBuilder) SelectArrayAggregation(alias string, source string, config ArrayAggConfig) SQLSelectChainBuilder {
	if config.Expr == "" {
		s.LastError = errors.New("expression should not empty")
	}

	var orderByClause string
	if config.SortBy != "" && config.SortOrder != 0 {
		order := "ASC"
		if config.SortOrder < 0 {
			order = "DESC"
		}
		orderByClause = fmt.Sprintf(" ORDER BY %s %s", config.SortBy, order)
	}

	if source != "" {
		source = " FROM " + source
	}
	arrayAggColumn := fmt.Sprintf(`(SELECT array_agg(%s%s)%s) AS "%s"`, config.Expr, orderByClause, source, alias)

	replaced := false
	for i, existing := range s.Columns {
		extracted := extractAlias(existing)
		if extracted != "" && extracted == strings.ToLower(alias) {
			s.Columns[i] = arrayAggColumn // Overwrite
			replaced = true
			break
		}
	}

	if !replaced {
		s.Columns = append(s.Columns, arrayAggColumn)
	}

	return s
}

func (s *SelectBuilder) SelectCaseWhen(thenExpr, elseExpr, alias string, whenClause string, whenArgs ...interface{}) SQLSelectChainBuilder {
	caseWhenColumn := fmt.Sprintf("CASE WHEN %s THEN %s ELSE %s END AS \"%s\"", whenClause, thenExpr, elseExpr, alias)

	// Check if alias exists in current list
	replaced := false
	for i, existing := range s.Columns {
		extracted := extractAlias(existing)
		if extracted != "" && extracted == strings.ToLower(alias) {
			s.Columns[i] = caseWhenColumn // Overwrite
			replaced = true
			break
		}
	}

	if !replaced {
		s.Columns = append(s.Columns, caseWhenColumn)
	}

	s.Args = append(s.Args, whenArgs...)
	return s
}

func (s *SelectBuilder) Join(table string, onCondition string, additionalConditions ...map[string]SQLCondition) SQLSelectChainBuilder {
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

func (s *SelectBuilder) LeftJoin(table string, mainCondition string, additionalConditions ...map[string]SQLCondition) SQLSelectChainBuilder {
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

func (s *SelectBuilder) LeftJoinLateralWithQuery(joinName string, joinQueryBuilder *SQLEloquentQuery, mainCondition string, additionalConditions ...map[string]SQLCondition) SQLSelectChainBuilder {
	joinQuery, joinArgs, err := joinQueryBuilder.Build()
	if err != nil {
		s.LastError = err
		return s
	}

	// Calculate the current argument offset
	offset := len(s.Args)

	// Shift the placeholders in the CTE query
	shiftedCTEQuery := shiftSQLPlaceholders(joinQuery, offset)

	s.Args = append(s.Args, joinArgs...)

	// additional filter
	var filterSb strings.Builder
	if len(additionalConditions) > 0 {
		var filters []string
		s.sharedWhereAndQuery(additionalConditions[0], &filters)

		for i := range filters {
			filterSb.WriteString(" AND ")
			filterSb.WriteString(filters[i])
		}
	}

	s.OtherTables = append(s.OtherTables, fmt.Sprintf("LEFT JOIN LATERAL (%s) %s ON %s%s", shiftedCTEQuery, joinName, mainCondition, filterSb.String()))
	return s
}

func (s *SelectBuilder) GroupBy(groupBy ...string) SQLSelectChainBuilder {
	s.Grouping = append(s.Grouping, groupBy...)
	return s
}

func (s *SelectBuilder) Having(havingClause map[string]SQLCondition) SQLSelectChainBuilder {
	s.useHaving = true
	s.SQLEloquentQuery.sharedWhereAndQuery(havingClause)
	return s
}

func (s *SelectBuilder) OrderBy(sortBy []string, asc bool) SQLSelectChainBuilder {
	direction := "ASC"
	nulls := "FIRST"

	if !asc {
		direction = "DESC"
		nulls = "LAST"
	}

	sortingRule := fmt.Sprintf("%s %s NULLS %s", strings.Join(sortBy, ", "), direction, nulls)
	s.SortBy = append(s.SortBy, sortingRule)
	return s
}

type Sort struct {
	SortBy    string `json:"sortBy"`
	SortOrder int    `json:"sortOrder"`
}

type SQLSort struct {
	SortBy    string `json:"sortBy"`
	SortOrder string `json:"sortOrder"` // ASC | DESC
}

type Pagination struct {
	Page        int    `json:"page"      transform:"int"`
	Limit       int    `json:"limit"     transform:"int"`
	SortBy      string `json:"sortBy"    transform:"string"`
	SortOrder   int    `json:"sortOrder" transform:"int"`
	MultiSort   []Sort `json:"multiSort"`
	DefaultSort []Sort `json:"defaultSort"`
}

func (s *SelectBuilder) Paginate(query Pagination) SQLSelectChainBuilder {
	var normalizedPage int
	if query.Page > 0 {
		normalizedPage = query.Page - 1
	} else {
		normalizedPage = 0
	}

	s.UsePagination = true
	s.Limit = query.Limit
	s.Offset = normalizedPage * query.Limit

	if query.SortBy != "" && query.SortOrder != 0 {
		s.OrderBy([]string{query.SortBy}, query.SortOrder > 0)
	} else if len(query.MultiSort) > 0 {
		// Overwrite sortBy sortOrder
		// Functions: For nameWithSequence sort (split the sort into 2 sorts)
		s.SortBy = []string{}

		for _, sort := range query.MultiSort {
			s.OrderBy([]string{sort.SortBy}, sort.SortOrder > 0)
		}
	} else if len(query.DefaultSort) > 0 {
		// Overwrite sortBy sortOrder
		// Functions: If all other sorts not defined, define defaultsort
		s.SortBy = []string{}

		for _, sort := range query.DefaultSort {
			s.OrderBy([]string{sort.SortBy}, sort.SortOrder > 0)
		}
	}

	return s
}

func (s *SelectBuilder) SetLimit(limit int) SQLSelectChainBuilder {
	var normalizedPage int

	s.Limit = limit
	s.Offset = normalizedPage * limit

	return s
}

func (s *SelectBuilder) WithCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLSelectChainBuilder {
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

func (s *SelectBuilder) WithRecursiveCTEBuilder(cteName string, cteBuilder *SQLEloquentQuery) SQLSelectChainBuilder {
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

func (s *SelectBuilder) UnionAll(cteBuilders ...*SQLEloquentQuery) SQLSelectChainBuilder {
	for _, cteBuilder := range cteBuilders {
		s.useUnionAll = true // only set true if len >0
		cteQuery, cteArgs, err := cteBuilder.Build()
		if err != nil {
			s.LastError = err
			return s
		}

		offset := len(s.Args)

		// Shift the placeholders in the CTE query
		shiftedQuery := shiftSQLPlaceholders(cteQuery, offset)

		s.UnionAllQueries = append(s.UnionAllQueries, shiftedQuery)
		s.Args = append(s.Args, cteArgs...)
	}

	return s
}

// NewSQLSelectBuilder creates a new chainable SELECT builder for a given table.
// It extracts JSON tags from the struct type T as default columns.
//
// Example:
//
//	builder := NewSQLSelectBuilder[User]("users","u").Select("id", "name")
func NewSQLSelectBuilder[T any](tableName string, alias ...string) SQLSelectChainBuilder {
	columns := ExtractJSONTags[T]()

	if len(alias) > 0 {
		tableName = fmt.Sprintf("%s %s", tableName, strings.TrimSpace(alias[0]))
	}

	defaultColumns := []string{}
	if len(columns) > 0 {
		defaultColumns = columns
	}
	return &SelectBuilder{
		&SQLEloquentQuery{
			Table:         tableName,
			Filters:       []string{},
			OtherTables:   []string{},
			Columns:       defaultColumns,
			Limit:         0,
			Offset:        0,
			SortBy:        []string{},
			CustomQuery:   "",
			Args:          nil,
			UsePagination: false,
			Mode:          "select",
		},
	}
}

func NewSQLSelectSubQueryBuilder[T any](tableName string, alias ...string) SQLSelectChainBuilder {
	columns := ExtractJSONTags[T]()

	if len(alias) > 0 {
		tableName = fmt.Sprintf("%s %s", tableName, strings.TrimSpace(alias[0]))
	}

	defaultColumns := []string{}
	if len(columns) > 0 {
		defaultColumns = columns
	}
	return &SelectBuilder{
		&SQLEloquentQuery{
			Table:         tableName,
			Filters:       []string{},
			OtherTables:   []string{},
			Columns:       defaultColumns,
			Limit:         0,
			Offset:        0,
			SortBy:        []string{},
			CustomQuery:   "",
			Args:          nil,
			UsePagination: false,
			Mode:          "select",
			isSubQuery:    true,
		},
	}
}

// Example:
//
//	builder := NewSQLCountBuilder("users","u").Where(...)
func NewSQLCountBuilder(tableName string, alias ...string) SQLSelectChainBuilder {
	defaultColumns := []string{"COUNT(*)"}

	if len(alias) > 0 {
		tableName = fmt.Sprintf("%s %s", tableName, strings.TrimSpace(alias[0]))
	}
	return &SelectBuilder{
		&SQLEloquentQuery{
			Table:           tableName,
			Filters:         []string{},
			OtherTables:     []string{},
			UnionAllQueries: []string{},
			Columns:         defaultColumns,
			Limit:           0,
			Offset:          0,
			SortBy:          []string{},
			CustomQuery:     "",
			Args:            nil,
			UsePagination:   false,
			Mode:            "select",
		},
	}
}

func (s *SQLEloquentQuery) buildSelectQuery() (string, []interface{}, error) {
	if len(s.HavingClauses) > 0 && len(s.Grouping) == 0 {
		return "", nil, errors.New("HAVING clauses only allowed if GROUP BY clause is exists")
	}

	if len(s.Columns) == 0 {
		s.Columns = []string{"*"}
	}

	var withSb strings.Builder
	var selectSb strings.Builder
	var joinSb strings.Builder
	var whereSb strings.Builder
	var groupSb strings.Builder
	var orderSb strings.Builder
	var havingSb strings.Builder
	withSb.Grow(256)   // preallocate ~256 bytes
	selectSb.Grow(256) // preallocate ~256 bytes
	joinSb.Grow(256)   // preallocate ~256 bytes
	whereSb.Grow(256)  // preallocate ~256 bytes
	groupSb.Grow(256)  // preallocate ~256 bytes
	orderSb.Grow(256)  // preallocate ~256 bytes
	havingSb.Grow(256) // preallocate ~256 bytes

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

	// DISTINCT
	var db strings.Builder
	if len(s.DistinctBy) > 0 {
		db.WriteString("DISTINCT ON ")
		db.WriteString("(")
		for i, d := range s.DistinctBy {
			if i > 0 {
				db.WriteString(",")
			}
			db.WriteString(d)
		}
		db.WriteString(")")
		db.WriteString(" " + s.DistinctAlias)

		tmp := make([]string, len(s.Columns))

		// Prepend distinct to the column clause.
		copy(tmp, s.Columns)
		s.Columns = []string{db.String()}
		s.Columns = append(s.Columns, tmp...)
	}

	// SELECT
	// Dont need to use default select if using union All, since union typically uses SELECT inside of it
	if !s.useUnionAll {
		selectSb.WriteByte('\n')
		selectSb.WriteString("SELECT ")
		for i, col := range s.Columns {
			if i > 0 {
				selectSb.WriteByte(',')
			}
			selectSb.WriteString(col)
		}
		selectSb.WriteByte('\n')
		selectSb.WriteString("FROM ")
		selectSb.WriteString(s.Table)
		selectSb.WriteByte('\n')
	} else if len(s.UnionAllQueries) > 0 { // UNION ALL
		for i, u := range s.UnionAllQueries {
			if i > 0 {
				selectSb.WriteString("UNION ALL")
				selectSb.WriteByte('\n')
			}
			selectSb.WriteString(u)
		}
		selectSb.WriteByte('\n')
	}

	// JOIN
	if len(s.OtherTables) > 0 {
		for _, j := range s.OtherTables {
			joinSb.WriteString(j)
			joinSb.WriteByte('\n')
		}
	}

	// WHERE
	if len(s.Filters) > 0 {
		whereSb.WriteString("WHERE ")
		for i, f := range s.Filters {
			if i > 0 {
				whereSb.WriteString(" AND ")
			}
			whereSb.WriteString(f)
		}
		whereSb.WriteByte('\n')
	}

	// GROUP BY
	if len(s.Grouping) > 0 {
		groupSb.WriteString("GROUP BY ")
		for i, g := range s.Grouping {
			if i > 0 {
				groupSb.WriteString(", ")
			}
			groupSb.WriteString(g)
		}
		groupSb.WriteByte('\n')
	}

	// HAVING
	if len(s.HavingClauses) > 0 {
		havingSb.WriteString("HAVING ")
		for i, h := range s.HavingClauses {
			if i > 0 {
				havingSb.WriteString(" AND ")
			}
			havingSb.WriteString(h)
		}
		havingSb.WriteByte('\n')
	}

	// ORDER BY
	// ORDER BY
	if len(s.SortBy) > 0 {
		orderSb.WriteString("ORDER BY ")

		asSplit := regexp.MustCompile(`(?i)\s+AS\s+`)
		stripDir := regexp.MustCompile(`\s+(?i)(asc|desc)\s*$`)

		// helper: trim space, trailing comma, and surrounding double quotes
		clean := func(s string) string {
			s = strings.TrimSpace(s)
			s = strings.TrimSuffix(s, ",")
			if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
				s = s[1 : len(s)-1]
			}
			return s
		}

		// map for transforming alias ("dataType") to column ("cf.data_type")
		aliasToExpr := make(map[string]string, len(s.Columns))
		for _, col := range s.Columns {
			col = strings.TrimSpace(col)
			parts := asSplit.Split(col, 2)

			if len(parts) == 2 {
				expr := strings.TrimSpace(parts[0]) // e.g. cf.data_type
				alias := clean(parts[1])            // e.g. "dataType" -> dataType
				key := strings.ToLower(alias)
				aliasToExpr[key] = expr
			} else {
				// support bare columns too (no AS)
				ident := clean(col)
				key := strings.ToLower(ident)
				aliasToExpr[key] = ident
			}
		}

		for i, srt := range s.SortBy {
			if i > 0 {
				orderSb.WriteString(", ")
			}

			srt = strings.TrimSpace(srt)

			// get dir if present (keep original casing)
			fields := strings.Fields(srt)
			dir := ""
			if n := len(fields); n > 1 && (strings.EqualFold(fields[n-1], "asc") || strings.EqualFold(fields[n-1], "desc")) {
				dir = " " + fields[n-1]
			}

			// key without ASC/DESC, then unquote/clean
			key := strings.TrimSpace(stripDir.ReplaceAllString(srt, ""))
			key = clean(key)
			lookup := strings.ToLower(key)

			// resolve alias -> expression (fallback to key as-is)
			if expr, ok := aliasToExpr[lookup]; ok {
				orderSb.WriteString(expr + dir)
			} else {
				orderSb.WriteString(key + dir)
			}
		}
		orderSb.WriteByte('\n')
	}

	// LIMIT/OFFSET
	if s.UsePagination {
		var limitationSb strings.Builder
		if s.Limit > 0 {
			limitationSb.WriteString("LIMIT ")
			limitationSb.WriteString(strconv.Itoa(s.Limit))
			limitationSb.WriteByte(' ')
		}
		limitationSb.WriteString("OFFSET ")
		limitationSb.WriteString(strconv.Itoa(s.Offset))
		limitationSb.WriteByte('\n')

		splittedTableName := strings.Split(s.Table, " ")
		prefix := splittedTableName[0]
		if len(splittedTableName) > 1 {
			prefix = splittedTableName[1]
		}

		mainQuery := selectSb.String() + joinSb.String() + fmt.Sprintf("JOIN paginated_ids ON paginated_ids.id = %s.id\n", prefix) + groupSb.String() + havingSb.String() + orderSb.String()
		filteredData := fmt.Sprintf("SELECT %s.id as id from %s\n", prefix, s.Table) + joinSb.String() + whereSb.String() + groupSb.String() + havingSb.String() + orderSb.String()
		paginatedDataQuery := "SELECT id as id from filtered_ids\n" + limitationSb.String()
		paginatedCountQuery := "SELECT COUNT(id) from filtered_ids\n"
		return PaginationQuery(withSb.String(), mainQuery, filteredData, paginatedDataQuery, paginatedCountQuery), s.Args, nil
	}

	query := withSb.String() + selectSb.String() + joinSb.String() + whereSb.String() + groupSb.String() + havingSb.String() + orderSb.String()
	return query, s.Args, nil
}

// Internal Utils for builder select
func extractAlias(column string) string {
	parts := strings.Split(strings.ToLower(column), " as ")
	if len(parts) < 2 {
		return "" // No alias
	}
	return strings.Trim(strings.TrimSpace(parts[len(parts)-1]), "\"")
}

// func validateGroupByColumns(availableColumns []string, groupByColumns []string) error {
// 	var columnNames []string
// 	for _, column := range availableColumns {
// 		columnNames = append(columnNames, extractAlias(column))
// 	}

// 	for _, each := range groupByColumns {
// 		if !util.ArrayIncludes(columnNames, strings.ToLower(each)) {
// 			return errors.New("GROUP BY columns should exist in selected columns")
// 		}
// 	}

// 	return nil
// }
