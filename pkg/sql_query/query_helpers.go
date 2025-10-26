package sql_query

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/mystaline/clefinport-be/pkg/dto"

	"github.com/jackc/pgx/v5"
)

func ScanRowObject(v any, row pgx.Rows) error {
	vVal := reflect.ValueOf(v)
	if vVal.Kind() != reflect.Ptr || vVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("ScanRow: v must be a pointer to a struct")
	}

	// Advance to first row
	if !row.Next() {
		if row.Err() != nil {
			return row.Err()
		}
		return pgx.ErrNoRows
	}

	fieldDescs := row.FieldDescriptions()

	values, err := row.Values()
	if err != nil {
		return err
	}

	rowMap := make(map[string]interface{})
	for i, fd := range fieldDescs {
		columnName := string(fd.Name)
		rowMap[columnName] = values[i]
	}

	jsonBytes, err := json.Marshal(rowMap)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, v); err != nil {
		return fmt.Errorf("unmarshal to struct failed: %w", err)
	}

	return nil
}

func CachedScanRowObject(v any, row pgx.Rows) error {
	vVal := reflect.ValueOf(v)
	if vVal.Kind() != reflect.Ptr || vVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("ScanRow: v must be a pointer to a struct")
	}

	if !row.Next() {
		if row.Err() != nil {
			return row.Err()
		}
		return pgx.ErrNoRows
	}

	fds := row.FieldDescriptions()
	values, err := row.Values()
	if err != nil {
		return err
	}

	fm := GetFieldMap(vVal.Type(), fds)
	structVal := vVal.Elem()

	for i, val := range values {
		idx := fm[i]
		if idx == -1 {
			continue
		}

		field := structVal.Field(idx)
		if !field.CanSet() {
			continue
		}

		rv := reflect.ValueOf(val)
		if rv.Type().AssignableTo(field.Type()) {
			field.Set(rv)
		} else if rv.Type().ConvertibleTo(field.Type()) {
			field.Set(rv.Convert(field.Type()))
		}
	}

	return nil
}

func ScanRowsArray(v any, rows pgx.Rows) error {
	vVal := reflect.ValueOf(v)
	if vVal.Kind() != reflect.Ptr || vVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("ScanRows: v must be a pointer to a slice")
	}

	sliceVal := vVal.Elem()
	elemType := sliceVal.Type().Elem()

	fieldDescs := rows.FieldDescriptions()
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return err
		}

		rowMap := make(map[string]interface{})
		for i, fd := range fieldDescs {
			columnName := string(fd.Name)
			rowMap[columnName] = values[i]
		}

		jsonBytes, err := json.Marshal(rowMap)
		if err != nil {
			return fmt.Errorf("marshal failed: %w", err)
		}

		newElemPtr := reflect.New(elemType) // *T
		if err := json.Unmarshal(jsonBytes, newElemPtr.Interface()); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
		}

		sliceVal.Set(reflect.Append(sliceVal, newElemPtr.Elem()))
	}

	return nil
}

func CachedScanRowsArray(v any, rows pgx.Rows) error {
	vVal := reflect.ValueOf(v)
	if vVal.Kind() != reflect.Ptr || vVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("ScanRows: v must be a pointer to a slice")
	}

	sliceVal := vVal.Elem()
	elemType := sliceVal.Type().Elem()
	fieldDescs := rows.FieldDescriptions()
	fm := GetFieldMap(elemType, fieldDescs)

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return err
		}

		newElemPtr := reflect.New(elemType)
		elem := newElemPtr.Elem()

		for i, val := range values {
			idx := fm[i]
			if idx == -1 {
				continue
			}

			field := elem.Field(idx)
			if !field.CanSet() {
				continue
			}

			rv := reflect.ValueOf(val)
			if rv.Type().AssignableTo(field.Type()) {
				field.Set(rv)
			} else if rv.Type().ConvertibleTo(field.Type()) {
				field.Set(rv.Convert(field.Type()))
			}
		}

		sliceVal.Set(reflect.Append(sliceVal, elem))
	}

	return nil
}

func FormatPaginationResult[T any](result []dto.PaginationResult[T]) dto.PaginationResult[T] {
	if len(result) < 1 {
		response := dto.PaginationResult[T]{
			Data:         []T{},
			TotalRecords: 0,
		}

		return response
	}

	return result[0]
}

/*
BuildInsertOneQuery generates an INSERT SQL query for a single struct.

Parameters:
- table: the name of the SQL table
- data: a struct with fields tagged with `json`

Returns:
- SQL query string
- list of arguments in correct order
- error if input is invalid

Example input:

	type MeasurementInput struct {
		Name string `json:"name"`
		Unit string `json:"unit"`
	}

	input := MeasurementInput{
		Name: "Height",
		Unit: "cm",
	}
*/
func BuildInsertOneQuery(table string, data interface{}) (query string, err error) {
	var args []interface{}

	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("data must be a struct")
	}

	t := v.Type()
	var columns []string
	var placeholders []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		columns = append(columns, tag)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
		args = append(args, v.Field(i).Interface())
	}

	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query, nil
}

/*
BuildInsertManyQuery generates an INSERT SQL query for a slice of structs.

Parameters:
- table: the name of the SQL table
- dataSlice: a slice of structs with fields tagged with `json`

Returns:
- SQL query string with multiple value sets
- list of arguments in correct order
- error if input is invalid

Example input:

	type MeasurementInput struct {
		Name string `json:"name"`
		Unit string `json:"unit"`
	}

	inputs := []MeasurementInput{
		{Name: "Height", Unit: "cm"},
		{Name: "Weight", Unit: "kg"},
		{Name: "Temp", Unit: "C"},
	}
*/
func BuildInsertManyQuery(table string, dataSlice interface{}) (query string, err error) {
	var args []interface{}

	v := reflect.ValueOf(dataSlice)

	if v.Kind() != reflect.Slice || v.Len() == 0 {
		return "", fmt.Errorf("data must be a non-empty slice of structs")
	}

	elemType := v.Index(0).Type()
	if elemType.Kind() != reflect.Struct {
		return "", fmt.Errorf("slice elements must be structs")
	}

	var columns []string
	for i := 0; i < elemType.NumField(); i++ {
		tag := elemType.Field(i).Tag.Get("json")
		if tag != "" && tag != "-" {
			columns = append(columns, tag)
		}
	}

	var valuePlaceholders []string
	argIndex := 1

	for i := 0; i < v.Len(); i++ {
		row := v.Index(i)
		var rowPlaceholders []string
		for j := 0; j < row.NumField(); j++ {
			field := elemType.Field(j)
			tag := field.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			args = append(args, row.Field(j).Interface())
			rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", argIndex))
			argIndex++
		}
		valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", ")))
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s RETURNING id",
		table,
		strings.Join(columns, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	return query, nil
}

func GenerateArrayFilter[T comparable](filter *[]T) SQLCondition {
	if filter == nil {
		var zeroVal T
		switch any(zeroVal).(type) {
		case string:
			return SQLCondition{
				Operator: SQLOperatorNotEqual,
				Value:    "__never_used_random_string__",
			}
		case int, int32, int64:
			return SQLCondition{
				Operator: SQLOperatorNotEqual,
				Value:    -999999999,
			}
		case float32, float64:
			return SQLCondition{
				Operator: SQLOperatorNotEqual,
				Value:    -99999999.99,
			}
		case bool:
			// return "column IN (true, false)" to match all rows
			return SQLCondition{
				Operator: SQLOperatorIn,
				Value:    []bool{true, false},
			}
		default:
			// fallback: column = column
			return SQLCondition{
				Operator: SQLOperatorEqual,
				Value:    zeroVal,
			}
		}
	}

	return SQLCondition{
		Operator: SQLOperatorIn,
		Value:    *filter,
	}
}

func PaginationQuery(withQuery, mainQuery, filteredDataQuery, paginatedDataQuery, paginatedCountQuery string) string {
	cleanWithQuery := strings.TrimPrefix(strings.TrimSpace(withQuery), "WITH")
	if cleanWithQuery != "" {
		cleanWithQuery = cleanWithQuery + ","
	}

	paginationQuery := fmt.Sprintf(`
		WITH
			%s
			filtered_ids AS (%s),
			paginated_ids AS (%s),
			total_query AS (%s),
			data_query AS (%s)
		SELECT
			COALESCE((SELECT jsonb_agg(data_query) FROM data_query), '[]') AS data,
			(SELECT COUNT FROM total_query) AS totalRecords;
	`, cleanWithQuery, filteredDataQuery, paginatedDataQuery, paginatedCountQuery, mainQuery)

	return paginationQuery
}

func PaginationQuery_old(dataQuery, totalQuery string) string {
	paginationQuery := fmt.Sprintf(`
		WITH
			paginated AS (%s),
			totalRow AS (%s),
			total AS (
				SELECT COUNT(*) FROM totalRow
			)
		SELECT
			COALESCE((SELECT jsonb_agg(paginated) FROM paginated), '[]') AS data,
			(SELECT COUNT FROM total) AS totalRecords;
	`, dataQuery, totalQuery)

	return paginationQuery
}

func ArrayIncludes[T comparable](slice []T, value T) bool {
	for _, each := range slice {
		if each == value {
			return true
		}
	}

	return false
}
