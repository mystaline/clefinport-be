package service

import (
	"context"
	"errors"
	"log"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/mystaline/clefinport-be/pkg/db"
	"github.com/mystaline/clefinport-be/pkg/dto"
	"github.com/mystaline/clefinport-be/pkg/sql_query"
	"github.com/mystaline/clefinport-be/pkg/sql_query/common_builders"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ReturningConfig struct {
	Column      []string
	Destination any
}

// Base Service PostgreSQL
type PostgreSqlService interface {
	// Debug sets the debug level for printing executed SQL queries.
	// Level:
	//
	//	0 → no logs
	//	1 → log queries
	//	2 → log queries and arguments (separated)
	//  3 → log queries and arguments (combined in the string)
	//
	// Defaults to level 1 if an invalid level is passed.
	// Remove its usage in production
	Debug(level ...int)
	// GetPool returns the underlying connection pool (PgxPoolInterface)
	// used by this service.
	GetPool() PgxPoolInterface
	// GetTransaction returns the underlying connection transaction (pgx.Tx)
	// used by this service.
	GetTransaction() pgx.Tx
	// SetTransaction assigns an active transaction (pgx.Tx) to the service.
	// If a transaction is set, all queries will be executed within it.
	SetTransaction(tx pgx.Tx)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error

	// Count executes a SELECT COUNT(*) query and returns the number of rows.
	Count(ctx context.Context, queryString string, args ...any) (int, error)
	// CountWithFilter builds a COUNT query using SQLCondition filters
	// and executes it.
	CountWithFilter(ctx context.Context, tableName string, filter map[string]sql_query.SQLCondition) (int, error)

	// Execute executes a query that wont return any value
	// Useful for defining temporary function and tables/views, so that the overall query is cleaner
	Execute(ctx context.Context, queryString string) error

	// SelectOne executes a SELECT query that returns a single row
	// and scans the result into the provided struct pointer v
	// (e.g., *dto.GetLoggedInUser).
	SelectOne(v any, ctx context.Context, queryString string, args ...any) error
	// SelectMany executes a SELECT query that returns multiple rows
	// and scans the results into the provided slice pointer v
	// (e.g., *[]dto.GetCustomFieldsResponse).
	SelectMany(v any, ctx context.Context, queryString string, args ...any) error

	// InsertOne executes an INSERT ... RETURNING id query
	// and returns the inserted row ID.
	InsertOne(ctx context.Context, queryString string, args ...any) (interface{}, error)
	// InsertOneWithData builds and executes an INSERT query for a single row,
	// using a struct or map (`body`) as the data source.
	//
	// Parameters:
	//
	//	ctx         - Context for timeout/cancellation propagation.
	//	body        - Struct or map representing columns and values to insert
	//	              (for struct, uses `column` tags).
	//	returnOption (optional) - Variadic ReturningConfig to specify:
	//	    • Column      - Slice of columns to include in RETURNING clause.
	//	    • Destination - Pointer to a struct to scan the inserted row.
	//
	// Behavior:
	//   - If returnOption is not provided → Executes INSERT ... RETURNING id, returns inserted ID.
	//   - If returnOption is provided with Destination → Executes INSERT ... RETURNING <columns>
	//     and scans the row into the given struct, returning nil as the first return value.
	//
	// Returns:
	//   - interface{} → Inserted ID (string) or nil (if Destination is provided).
	//   - error       → Any error encountered during query building, execution, or scanning.
	InsertOneWithData(
		ctx context.Context,
		tableName string,
		body interface{},
		returnOption ...ReturningConfig,
	) (interface{}, error)
	// InsertMany executes an INSERT query for multiple rows
	// and returns the number of rows affected.
	InsertMany(ctx context.Context, queryString string, args ...any) (int64, error)
	// InsertManyWithData builds and executes an INSERT query for multiple rows,
	// using a slice of structs or maps (`body`) as the data source.
	//
	// Parameters:
	//
	//	ctx         - Context for timeout/cancellation propagation.
	//	body        - Slice of structs or maps representing rows to insert.
	//	returnOption (optional) - Variadic ReturningConfig to specify:
	//	    • Column      - Slice of columns to include in RETURNING clause.
	//	    • Destination - Pointer to a slice to scan all inserted rows.
	//
	// Behavior:
	//   - If returnOption is not provided → Executes INSERT and returns rows affected.
	//   - If returnOption is provided with Destination → Executes INSERT ... RETURNING <columns>,
	//     scans the rows into Destination, and returns the length of the slice.
	//
	// Returns:
	//   - interface{} → Either int64 (rows affected) or int64 (len of scanned rows).
	//   - error       → Any error encountered during query building, execution, or scanning.
	//
	// Notes:
	//   - Destination must be a pointer to a slice (e.g., *[]YourDTO), otherwise it returns an error.
	InsertManyWithData(
		ctx context.Context,
		tableName string,
		body interface{},
		returnOption ...ReturningConfig,
	) (interface{}, error)

	// UpdateOne executes an UPDATE ... RETURNING id query
	// and returns the updated row ID.
	UpdateOne(ctx context.Context, queryString string, args ...any) (interface{}, error)
	// UpdateOneWithData builds and executes an UPDATE query using a filter map (`query`)
	// and a struct or map as the update body (`body`).
	//
	// Parameters:
	//
	//	ctx         - The request context for cancellation and deadlines.
	//	query       - A map of column names to SQLCondition objects used in the WHERE clause.
	//	body        - A struct or map representing the columns to update (uses `column` tags for structs).
	//	returnOption (optional) - A variadic argument of ReturningConfig that defines:
	//	    • Column      - A slice of column names to include in the RETURNING clause.
	//	    • Destination - A pointer to a struct where the returned row will be scanned.
	//	                    If Destination is nil, the method returns the updated row's ID as string.
	//
	// Behavior:
	//   - If returnOption is not provided → executes UPDATE ... RETURNING id and returns the ID (string).
	//   - If returnOption is provided but Destination is nil → still executes UPDATE and returns the ID.
	//   - If returnOption is provided with Destination → scans the updated row into Destination
	//     and returns nil as the first return value (caller uses Destination for data).
	//
	// Returns:
	//   - interface{}: The updated row's ID (string) if no Destination is provided.
	//     nil if Destination is provided (caller reads from Destination).
	//   - error:       Any error encountered during query building, execution, or scanning.
	UpdateOneWithData(
		ctx context.Context,
		tableName string,
		query map[string]sql_query.SQLCondition,
		body interface{},
		returnOption ...ReturningConfig,
	) (interface{}, error)
	// UpdateMany executes an UPDATE query that may affect multiple rows
	// and returns the number of rows updated.
	UpdateMany(ctx context.Context, queryString string, args ...any) (int64, error)
	// UpdateManyWithData builds and executes an UPDATE query that may affect multiple rows.
	// It supports returning and scanning the updated rows into a destination slice.
	//
	// Parameters:
	//
	//	ctx         - The request context for cancellation and deadlines.
	//	query       - A map of column names to SQLCondition objects used in the WHERE clause.
	//	body        - A struct or map representing the columns to update (uses `column` tags for structs).
	//	returnOption (optional) - A variadic argument of ReturningConfig that defines:
	//	    • Column      - Slice of column names to include in RETURNING.
	//	    • Destination - Pointer to a slice where all updated rows will be scanned.
	//
	// Behavior:
	//   - If no returnOption is provided → executes UPDATE and returns the number of rows affected.
	//   - If returnOption is provided with Destination → executes UPDATE ... RETURNING columns
	//     and scans all updated rows into Destination (which must be a pointer to a slice).
	//     The returned int64 equals the length of the slice.
	//   - If Destination is not a pointer to a slice → returns an error.
	//
	// Returns:
	//   - int64: Number of rows affected. If Destination is provided, this equals len(Destination).
	//   - error: Any error encountered during query building, execution, or scanning.
	UpdateManyWithData(
		ctx context.Context,
		tableName string,
		query map[string]sql_query.SQLCondition,
		body interface{},
		returnOption ...ReturningConfig,
	) (int64, error)
	// UpdateEachWithData performs a bulk update per row (row-specific values)
	// using rowIdentifier (typically a primary key column or unique index).
	UpdateEachWithData(
		ctx context.Context,
		tableName string,
		rowIdentifier string,
		query map[string]sql_query.SQLCondition,
		body interface{},
	) (int64, error)

	// SoftDeleteOne builds and executes a UPDATE soft delete query for a single row
	// using SQLCondition filters and returns the deleted row ID.
	// Table must has column is_deleted and deleted_at.
	SoftDeleteOne(
		ctx context.Context,
		tableName string,
		filter map[string]sql_query.SQLCondition,
		returnOption ...ReturningConfig,
	) (interface{}, error)
	// SoftDeleteMany builds and executes a UPDATE soft delete query for multiple rows
	// using SQLCondition filters and returns the number of rows deleted.
	// Table must has column is_deleted and deleted_at.
	SoftDeleteMany(
		ctx context.Context,
		tableName string,
		filter map[string]sql_query.SQLCondition,
		returnOption ...ReturningConfig,
	) (int64, error)

	// DeleteOne executes a DELETE ... RETURNING id query
	// and returns the deleted row ID.
	DeleteOne(ctx context.Context, queryString string, args ...any) (interface{}, error)
	// DeleteOneWithFilter builds and executes a DELETE query for a single row
	// using SQLCondition filters and returns the deleted row ID.
	DeleteOneWithFilter(
		ctx context.Context,
		tableName string,
		filter map[string]sql_query.SQLCondition,
	) (interface{}, error)
	// DeleteMany executes a DELETE query that may affect multiple rows
	// and returns the number of rows deleted.
	DeleteMany(ctx context.Context, queryString string, args ...any) (int64, error)
	// DeleteManyWithFilter builds and executes a DELETE query for multiple rows
	// using SQLCondition filters and returns the number of rows deleted.
	DeleteManyWithFilter(
		ctx context.Context,
		tableName string,
		filter map[string]sql_query.SQLCondition,
	) (int64, error)
}

type PgxPoolInterface interface {
	// Begin starts a new database transaction and returns a pgx.Tx.
	Begin(ctx context.Context) (pgx.Tx, error)

	// Query executes a SQL query and returns pgx.Rows for multiple row results.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	// QueryRow executes a SQL query that returns a single row.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	// Exec executes a SQL command (INSERT/UPDATE/DELETE)
	// and returns a pgconn.CommandTag containing rows affected.
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)

	CopyFrom(ctx context.Context, identifier pgx.Identifier, columns []string, source pgx.CopyFromSource) (int64, error)

	// Close closes the database pool connection.
	Close()
}

type BasePostgreSqlService struct {
	Pool        PgxPoolInterface
	Transaction pgx.Tx

	debugLevel int
}

// MakeService creates a new PostgreSqlService instance.
func MakeService(dbName db.DBName) PostgreSqlService {
	pool := db.ConnectPostgres(dbName)

	return &BasePostgreSqlService{Pool: pool}
}

func (s *BasePostgreSqlService) Debug(level ...int) {
	if len(level) > 0 && sql_query.ArrayIncludes([]int{1, 2, 3}, level[0]) {
		s.debugLevel = level[0]
		return
	}

	s.debugLevel = 1
}

func (s *BasePostgreSqlService) GetPool() PgxPoolInterface {
	return s.Pool
}

func (s *BasePostgreSqlService) GetTransaction() pgx.Tx {
	return s.Transaction
}

func (s *BasePostgreSqlService) SetTransaction(tx pgx.Tx) {
	s.Transaction = tx
}

func (s *BasePostgreSqlService) CommitTransaction(ctx context.Context) error {
	if s.Transaction == nil {
		return errors.New("no active transaction to commit")
	}

	if err := s.Transaction.Commit(ctx); err != nil {
		_ = s.Transaction.Rollback(ctx)
		return err
	}

	return nil
}

func (s *BasePostgreSqlService) RollbackTransaction(ctx context.Context) error {
	if s.Transaction == nil {
		return errors.New("no active transaction to rollback")
	}

	_ = s.Transaction.Rollback(ctx)
	return nil
}

func (s *BasePostgreSqlService) Count(
	ctx context.Context,
	queryString string,
	args ...any,
) (int, error) {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var count int
	var err error

	if s.Transaction != nil {
		err = s.Transaction.QueryRow(ctx, queryString, args...).Scan(&count)
	} else {
		err = s.Pool.QueryRow(ctx, queryString, args...).Scan(&count)
	}

	if err != nil {
		log.Println("Count query failed:", err)
		return 0, err
	}

	return count, nil
}

func (s *BasePostgreSqlService) Execute(
	ctx context.Context,
	queryString string,
) error {
	shouldShowQuery(s.debugLevel, queryString)

	var rows pgx.Rows
	var err error

	if s.Transaction != nil {
		rows, err = s.Transaction.Query(ctx, queryString)
	} else {
		rows, err = s.Pool.Query(ctx, queryString)
	}

	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func (s *BasePostgreSqlService) CountWithFilter(
	ctx context.Context,
	tableName string,
	filter map[string]sql_query.SQLCondition,
) (int, error) {
	queryString, args := common_builders.CountBuilder(tableName, filter)

	return s.Count(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) SelectOne(
	v any,
	ctx context.Context,
	queryString string,
	args ...any,
) error {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var rows pgx.Rows
	var err error

	if s.Transaction != nil {
		rows, err = s.Transaction.Query(ctx, queryString, args...)
	} else {
		rows, err = s.Pool.Query(ctx, queryString, args...)
	}

	if err != nil {
		return err
	}
	defer rows.Close()

	err = sql_query.ScanRowObject(v, rows)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *BasePostgreSqlService) SelectMany(
	v any,
	ctx context.Context,
	queryString string,
	args ...any,
) error {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var rows pgx.Rows
	var err error

	if s.Transaction != nil {
		rows, err = s.Transaction.Query(ctx, queryString, args...)
	} else {
		rows, err = s.Pool.Query(ctx, queryString, args...)
	}

	if err != nil {
		return err
	}
	defer rows.Close()

	err = sql_query.ScanRowsArray(v, rows)
	if err != nil {
		log.Println(err)
		return err
	}

	if rows.Err() != nil {
		log.Printf("rows error: %v", rows.Err())
		return rows.Err()
	}

	return nil
}

func (s *BasePostgreSqlService) InsertOne(
	ctx context.Context,
	queryString string,
	args ...any,
) (interface{}, error) {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var resultId int
	var err error

	if s.Transaction != nil {
		err = s.Transaction.QueryRow(ctx, queryString, args...).Scan(&resultId)
	} else {
		err = s.Pool.QueryRow(ctx, queryString, args...).Scan(&resultId)
	}

	if err != nil {
		return nil, err
	}

	return resultId, nil
}

func (s *BasePostgreSqlService) InsertOneWithData(
	ctx context.Context,
	tableName string,
	body interface{},
	returnOption ...ReturningConfig,
) (interface{}, error) {
	returnColumn := []string{}

	if len(returnOption) > 0 {
		returnColumn = append(returnColumn, returnOption[0].Column...)
	}
	queryString, args := common_builders.InsertBuilder(tableName, body, returnColumn...)

	if len(returnOption) > 0 && returnOption[0].Destination != nil {
		return nil, s.SelectOne(returnOption[0].Destination, ctx, queryString, args...)
	}
	return s.InsertOne(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) InsertMany(
	ctx context.Context,
	queryString string,
	args ...any,
) (int64, error) {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var commandTag pgconn.CommandTag
	var err error

	if s.Transaction != nil {
		commandTag, err = s.Transaction.Exec(ctx, queryString, args...)
	} else {
		commandTag, err = s.Pool.Exec(ctx, queryString, args...)
	}

	if err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}

func (s *BasePostgreSqlService) InsertManyWithData(
	ctx context.Context,
	tableName string,
	body interface{},
	returnOption ...ReturningConfig,
) (interface{}, error) {
	returnColumn := []string{}

	if len(returnOption) > 0 {
		returnColumn = append(returnColumn, returnOption[0].Column...)
	}
	queryString, args := common_builders.InsertBuilder(tableName, body, returnColumn...)

	if len(returnOption) > 0 && returnOption[0].Destination != nil {
		err := s.SelectMany(returnOption[0].Destination, ctx, queryString, args...)
		val := reflect.ValueOf(returnOption[0].Destination)

		return int64(val.Elem().Len()), err
	}
	return s.InsertMany(ctx, queryString, args...)
}

// Still in experimental stage, recommended to use InsertManyWithData until this function stable
func (s *BasePostgreSqlService) InsertBatch(
	ctx context.Context,
	tableName string,
	body interface{},
) (int64, error) {
	v := reflect.ValueOf(body)
	if v.Kind() != reflect.Slice {
		return 0, errors.New("body must be a slice")
	}
	if v.Len() == 0 {
		return 0, nil
	}

	firstElem := v.Index(0).Interface()
	t := reflect.TypeOf(firstElem)
	typeName := t.PkgPath() + "." + t.Name()

	// Ambil atau buat template
	insertTemplate, ok := sql_query.InsertCache[typeName]
	if !ok {
		insertTemplate = sql_query.BuildInsertTemplate(t) // disini tentukan useNow & useID
		sql_query.InsertCache[typeName] = insertTemplate
	}

	rows := make([][]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		row := make([]interface{}, len(insertTemplate.InsertColumn))

		for j := range insertTemplate.InsertColumn {
			if insertTemplate.UseID[j] {
				row[j] = db.Node.Generate().Int64()
			} else if insertTemplate.UseNow[j] {
				row[j] = time.Now() // CopyFrom tidak bisa pakai literal NOW(), harus diisi value
			} else {
				row[j] = elem.FieldByIndex(insertTemplate.FieldIndexes[j]).Interface()
			}
		}
		rows[i] = row
	}

	if s.Transaction != nil {
		return s.Transaction.CopyFrom(
			ctx,
			pgx.Identifier{tableName},
			insertTemplate.InsertColumn,
			pgx.CopyFromRows(rows),
		)
	}

	return s.Pool.CopyFrom(
		ctx,
		pgx.Identifier{tableName},
		insertTemplate.InsertColumn,
		pgx.CopyFromRows(rows),
	)
}

func (s *BasePostgreSqlService) UpdateOne(
	ctx context.Context,
	queryString string,
	args ...any,
) (interface{}, error) {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var resultId int
	var err error

	if s.Transaction != nil {
		err = s.Transaction.QueryRow(ctx, queryString, args...).Scan(&resultId)
	} else {
		err = s.Pool.QueryRow(ctx, queryString, args...).Scan(&resultId)
	}

	if err != nil {
		return nil, err
	}

	return resultId, nil
}

func (s *BasePostgreSqlService) UpdateOneWithData(
	ctx context.Context,
	tableName string,
	query map[string]sql_query.SQLCondition,
	body interface{},
	returnOption ...ReturningConfig,
) (interface{}, error) {
	returnColumn := []string{}

	if len(returnOption) > 0 {
		returnColumn = append(returnColumn, returnOption[0].Column...)
	}
	queryString, args := common_builders.UpdateBuilder(tableName,
		query,
		body,
		returnColumn...,
	)

	if len(returnOption) > 0 && returnOption[0].Destination != nil {
		return nil, s.SelectOne(returnOption[0].Destination, ctx, queryString, args...)
	}
	return s.UpdateOne(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) UpdateMany(
	ctx context.Context,
	queryString string,
	args ...any,
) (int64, error) {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var commandTag pgconn.CommandTag
	var err error

	if s.Transaction != nil {
		commandTag, err = s.Transaction.Exec(ctx, queryString, args...)
	} else {
		commandTag, err = s.Pool.Exec(ctx, queryString, args...)
	}

	if err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}

func (s *BasePostgreSqlService) UpdateManyWithData(
	ctx context.Context,
	tableName string,
	query map[string]sql_query.SQLCondition,
	body interface{},
	returnOption ...ReturningConfig,
) (int64, error) {
	returnColumn := []string{}

	if len(returnOption) > 0 {
		returnColumn = append(returnColumn, returnOption[0].Column...)
	}
	queryString, args := common_builders.UpdateBuilder(tableName,
		query,
		body,
		returnColumn...,
	)

	if len(returnOption) > 0 && returnOption[0].Destination != nil {
		err := s.SelectMany(returnOption[0].Destination, ctx, queryString, args...)
		val := reflect.ValueOf(returnOption[0].Destination)

		return int64(val.Elem().Len()), err
	}
	return s.UpdateMany(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) UpdateEachWithData(
	ctx context.Context,
	tableName string,
	rowIdentifier string,
	query map[string]sql_query.SQLCondition,
	body interface{},
) (int64, error) {
	queryString, args := common_builders.UpdateEachBuilder(tableName,
		rowIdentifier,
		query,
		body,
	)

	return s.UpdateMany(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) SoftDeleteOne(
	ctx context.Context,
	tableName string,
	filter map[string]sql_query.SQLCondition,
	returnOption ...ReturningConfig,
) (interface{}, error) {
	returnColumn := []string{}

	if len(returnOption) > 0 {
		returnColumn = append(returnColumn, returnOption[0].Column...)
	}
	queryString, args := common_builders.UpdateBuilder(tableName, filter, dto.SetSoftDelete{
		IsDeleted: true,
		DeletedAt: "NOW()",
	}, returnColumn...)
	shouldShowQuery(s.debugLevel, queryString, args...)

	if len(returnOption) > 0 && returnOption[0].Destination != nil {
		return nil, s.SelectOne(returnOption[0].Destination, ctx, queryString, args...)
	}
	return s.DeleteOne(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) SoftDeleteMany(
	ctx context.Context,
	tableName string,
	filter map[string]sql_query.SQLCondition,
	returnOption ...ReturningConfig,
) (int64, error) {
	returnColumn := []string{}

	if len(returnOption) > 0 {
		returnColumn = append(returnColumn, returnOption[0].Column...)
	}
	queryString, args := common_builders.UpdateBuilder(tableName,
		filter,
		dto.SetSoftDelete{
			IsDeleted: true,
			DeletedAt: "NOW()",
		},
		returnColumn...,
	)
	shouldShowQuery(s.debugLevel, queryString, args...)

	if len(returnOption) > 0 && returnOption[0].Destination != nil {
		err := s.SelectMany(returnOption[0].Destination, ctx, queryString, args...)
		val := reflect.ValueOf(returnOption[0].Destination)

		return int64(val.Elem().Len()), err
	}
	return s.DeleteMany(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) DeleteOne(
	ctx context.Context,
	queryString string,
	args ...any,
) (interface{}, error) {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var resultId int
	var err error

	if s.Transaction != nil {
		err = s.Transaction.QueryRow(ctx, queryString, args...).Scan(&resultId)
	} else {
		err = s.Pool.QueryRow(ctx, queryString, args...).Scan(&resultId)
	}

	if err != nil {
		return nil, err
	}

	return resultId, nil
}

func (s *BasePostgreSqlService) DeleteOneWithFilter(
	ctx context.Context,
	tableName string,
	filter map[string]sql_query.SQLCondition,
) (interface{}, error) {
	queryString, args := common_builders.DeleteBuilder(tableName, filter)

	return s.DeleteOne(ctx, queryString, args...)
}

func (s *BasePostgreSqlService) DeleteMany(
	ctx context.Context,
	queryString string,
	args ...any,
) (int64, error) {
	shouldShowQuery(s.debugLevel, queryString, args...)

	var commandTag pgconn.CommandTag
	var err error

	if s.Transaction != nil {
		commandTag, err = s.Transaction.Exec(ctx, queryString, args...)
	} else {
		commandTag, err = s.Pool.Exec(ctx, queryString, args...)
	}

	if err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}

func (s *BasePostgreSqlService) DeleteManyWithFilter(
	ctx context.Context,
	tableName string,
	filter map[string]sql_query.SQLCondition,
) (int64, error) {
	queryString, args := common_builders.DeleteBuilder(tableName, filter)

	return s.DeleteMany(ctx, queryString, args...)
}

// UseTransactions executes fn within a transaction.
// If fn returns an error, the transaction is rolled back.
// If fn succeeds, the transaction is committed.
// The returned value is whatever fn returns, or an error if commit/rollback fails.
//
// ⚠️ WARNING about holdCommit mode:
// When holdCommit is true, this function will NOT commit or rollback.
// Instead, the caller is fully responsible for calling tx.Rollback or tx.Commit.
// Forgetting to do so will:
//   - Leave the connection "dirty" and unusable for other queries
//   - Hold locks and temporary resources open in PostgreSQL
//   - Eventually cause connection pool exhaustion under load
//
// Although pgx will automatically rollback when the connection is returned
// to the pool, you must NOT rely on this. Always explicitly close the tx.
func UseTransactions[T any](
	ctx context.Context,
	pool PgxPoolInterface,
	fn func(tx pgx.Tx) (T, error),
	holdCommit ...bool,
) (result T, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("can't start transactions: %v", err)
		err = errors.New("something went wrong")
		return
	}

	// When holdCommit is true, caller is responsible for calling Rollback or Commit. Transaction is returned unclosed.
	if len(holdCommit) > 0 && holdCommit[0] {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("transaction panicked: %v\n%s", r, debug.Stack())
				err = errors.New("something went wrong")
			}
		}()

		result, err = fn(tx)

		if err != nil {
			log.Printf("Error %v\n", err.Error())
		}

		return result, err
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("transaction panicked: %v\n%s", r, debug.Stack())
			_ = tx.Rollback(ctx)
			err = errors.New("something went wrong")
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	result, err = fn(tx)
	if err != nil {
		log.Printf("Error %v\n", err.Error())
		return
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		log.Printf("failed to commit: %v", commitErr)
		err = errors.New("something went wrong")
		return
	}

	return result, nil
}

func shouldShowQuery(level int, query string, args ...any) {
	switch level {
	case 1:
		log.Println("query", query)
	case 2:
		log.Println("query and args", query, args)
	}
}
