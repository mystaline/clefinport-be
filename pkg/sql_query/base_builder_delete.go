package sql_query

import (
	"errors"
	"fmt"
	"strings"
)

// To ensure SQLDeleteBuilder has its own initial methods
// e.g. DeleteBuilder(...).Delete()...Rest
type SQLDeleteInitBuilder interface {
	// Delete implements SQLDeleteChainBuilder. (Only able to be called once)
	// Delete initializes a DELETE statement for the current table.
	// By default, it returns the deleted "id", but you can pass custom RETURNING columns.
	//
	// Example:
	//
	//	builder.Delete()                     // RETURNING id
	//	builder.Delete("id", "deleted_at")   // RETURNING id, deleted_at
	//
	// Generates:
	//
	//	DELETE FROM table_name RETURNING id
	Delete(returningColumns ...string) SQLDeleteChainBuilder
}

// To ensure method .Delete() has its own chaining methods
// e.g. .Delete(...).Using(...).Where(...).Build()
type SQLDeleteChainBuilder interface {
	// Where implements SQLDeleteChainBuilder. (Accumulates previous value if called again)
	Where(filters map[string]SQLCondition) SQLDeleteChainBuilder
	// WhereOr implements SQLDeleteChainBuilder. (Accumulates previous value if called again)
	WhereOr(filters ...map[string]SQLCondition) SQLDeleteChainBuilder

	// Using implements SQLDeleteChainBuilder. (Overrides previous value if called again)
	// Using adds a USING clause to the DELETE statement.
	// It is useful for multi-table DELETE with a join-like behavior.
	//
	// Example:
	//
	//	builder.Using([]string{"other_table"})
	//
	// Generates:
	//
	//	DELETE FROM table_name USING other_table
	Using(tables []string) SQLDeleteChainBuilder

	// buildDeleteQuery finalizes the DELETE query into a full SQL string + args.
	// It adds USING, WHERE, and RETURNING clauses if provided.
	// Prevents execution if CustomQuery is empty.
	//
	// Example Output:
	//
	//	DELETE FROM users USING roles WHERE users.role_id = roles.id RETURNING id
	Build() (string, []interface{}, error)
}

// <---To wrap builder to its respective interface, used as pointer type in method of each builder--->
type DeleteBuilder struct {
	*SQLEloquentQuery
}

func (s *DeleteBuilder) Where(filters map[string]SQLCondition) SQLDeleteChainBuilder {
	s.SQLEloquentQuery.sharedWhereAndQuery(filters)
	return s
}

func (s *DeleteBuilder) WhereOr(filters ...map[string]SQLCondition) SQLDeleteChainBuilder {
	s.SQLEloquentQuery.sharedWhereOrQuery(filters...)
	return s
}

func (s *DeleteBuilder) Delete(returningColumns ...string) SQLDeleteChainBuilder {
	if len(returningColumns) > 0 {
		s.Columns = returningColumns
	} else {
		s.Columns = []string{"id"}
	}

	s.CustomQuery = fmt.Sprintf(
		"DELETE FROM %s",
		s.Table,
	)
	return s
}

func (s *DeleteBuilder) Using(tables []string) SQLDeleteChainBuilder {
	if len(tables) < 1 {
		return s
	}

	var otherTables []string
	otherTables = append(otherTables, fmt.Sprintf("USING %s", strings.Join(tables, ", ")))
	s.OtherTables = otherTables
	return s
}

// NewSQLDeleteBuilder creates a new delete builder for a given table.
//
// Example:
//
//	builder := NewSQLDeleteBuilder("users","u")
func NewSQLDeleteBuilder(tableName string, alias ...string) SQLDeleteInitBuilder {
	if len(alias) > 0 {
		tableName = fmt.Sprintf("%s %s", tableName, strings.TrimSpace(alias[0]))
	}

	return &DeleteBuilder{
		&SQLEloquentQuery{
			Table:       tableName,
			Filters:     []string{},
			OtherTables: []string{},
			Columns:     []string{},
			CustomQuery: "",
			Args:        nil,
			Mode:        "delete",
		},
	}
}

func (s *SQLEloquentQuery) buildDeleteQuery() (string, []interface{}, error) {
	if s.LastError != nil {
		return "", nil, errors.New(s.LastError.Error())
	}

	if s.CustomQuery == "" {
		return "", nil, errors.New("invalid update query: CustomQuery not set")
	}

	query := s.CustomQuery

	if len(s.OtherTables) > 0 {
		usingClause := strings.Join(s.OtherTables, " ")
		query += " " + usingClause
	}

	if len(s.Filters) > 0 {
		query += " " + fmt.Sprintf("WHERE %s", strings.Join(s.Filters, " AND "))
	}

	if len(s.Columns) > 0 {
		query += " RETURNING " + strings.Join(s.Columns, ",")
	}

	return query, s.Args, nil
}
