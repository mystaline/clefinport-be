package sql_query

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func (s *SelectBuilder) SelectJSONArrayElements(alias string, arrayElements []map[string]string) SQLSelectChainBuilder {
	if alias == "" {
		alias = "array_elements"
	}

	jsonBytes, err := json.Marshal(arrayElements)
	if err != nil {
		s.LastError = fmt.Errorf("SelectJSONArrayElements marshal error: %w", err)
		return s
	}

	jsonStr := string(jsonBytes)
	placeholder := len(s.Args) + 1
	s.Args = append(s.Args, jsonStr)

	formatted := fmt.Sprintf(`jsonb_array_elements(%d::jsonb) AS "%s"`, placeholder, alias)
	s.Columns = append(s.Columns, formatted)

	return s
}

func (s *SelectBuilder) SelectJSONAggregate(alias string, dto any, condition string, asArrayAggregation bool, orderByClauses ...string) SQLSelectChainBuilder {
	var mappedJSON map[string]string

	v := reflect.ValueOf(dto)
	if v.Kind() != reflect.Struct {
		mappedJSON = dto.(map[string]string)
	} else {
		mappedJSON = MakeMapJSONTagsFromValue(dto)
	}

	if len(mappedJSON) == 0 {
		return s
	}

	var keyValuePairs []string
	var keys []string
	for k := range mappedJSON {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, jsonKey := range keys {
		columnExpr := mappedJSON[jsonKey]
		keyValuePairs = append(keyValuePairs,
			fmt.Sprintf("'%s'", jsonKey),
			columnExpr,
		)
	}

	var formattedColumn string
	if asArrayAggregation {
		orderBy := ""
		if len(orderByClauses) > 0 {
			orderBy = fmt.Sprintf(" ORDER BY %s", orderByClauses[0])
		}
		formattedColumn = fmt.Sprintf("jsonb_agg(jsonb_build_object(%s)%s)", strings.Join(keyValuePairs, ", "), orderBy)
		if condition != "" {
			formattedColumn = fmt.Sprintf("%s FILTER (WHERE %s)", formattedColumn, condition)
		}
	} else {
		formattedColumn = fmt.Sprintf("jsonb_build_object(%s)", strings.Join(keyValuePairs, ", "))
		if condition != "" {
			formattedColumn = fmt.Sprintf("CASE WHEN %s THEN %s ELSE NULL END", condition, formattedColumn)
		}
	}
	formattedColumn = fmt.Sprintf(`%s AS "%s"`, formattedColumn, alias)

	if s.WrapAggregation && !asArrayAggregation {
		s.NestedAggregation = append(s.NestedAggregation, formattedColumn)
	} else {
		s.Columns = append(s.Columns, formattedColumn)
	}

	return s
}

func (s *SelectBuilder) SelectJSONAggregateCoalesce(alias string, dto any, condition string, asArrayAggregation bool, coalesce string, orderByClauses ...string) SQLSelectChainBuilder {
	var mappedJSON map[string]string

	v := reflect.ValueOf(dto)
	if v.Kind() != reflect.Struct {
		mappedJSON = dto.(map[string]string)
	} else {
		mappedJSON = MakeMapJSONTagsFromValue(dto)
	}

	if len(mappedJSON) == 0 {
		return s
	}

	var keyValuePairs []string
	var keys []string
	for k := range mappedJSON {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, jsonKey := range keys {
		columnExpr := mappedJSON[jsonKey]
		keyValuePairs = append(keyValuePairs,
			fmt.Sprintf("'%s'", jsonKey),
			columnExpr,
		)
	}

	var formattedColumn string
	if asArrayAggregation {
		orderBy := ""
		if len(orderByClauses) > 0 {
			orderBy = fmt.Sprintf(" ORDER BY %s", orderByClauses[0])
		}
		formattedColumn = fmt.Sprintf("jsonb_agg(jsonb_build_object(%s)%s)", strings.Join(keyValuePairs, ", "), orderBy)

		if condition != "" {
			formattedColumn = fmt.Sprintf("%s FILTER (WHERE %s)", formattedColumn, condition)
		}
		formattedColumn = fmt.Sprintf("COALESCE(%s,%s)", formattedColumn, coalesce)

	} else {
		formattedColumn = fmt.Sprintf("jsonb_build_object(%s)", strings.Join(keyValuePairs, ", "))

		if condition != "" {
			formattedColumn = fmt.Sprintf("CASE WHEN %s THEN %s ELSE NULL END", condition, formattedColumn)
		}
		formattedColumn = fmt.Sprintf("COALESCE(%s,%s)", formattedColumn, coalesce)

	}
	formattedColumn = fmt.Sprintf(`%s AS "%s"`, formattedColumn, alias)

	if s.WrapAggregation && !asArrayAggregation {
		s.NestedAggregation = append(s.NestedAggregation, formattedColumn)
	} else {
		s.Columns = append(s.Columns, formattedColumn)
	}

	return s
}

func (s *SelectBuilder) SelectJSONAggregateDistinct(alias string, dto any, condition string, asArrayAggregation bool, orderByClauses ...string) SQLSelectChainBuilder {
	var mappedJSON map[string]string

	v := reflect.ValueOf(dto)
	if v.Kind() != reflect.Struct {
		mappedJSON = dto.(map[string]string)
	} else {
		mappedJSON = MakeMapJSONTagsFromValue(dto)
	}

	if len(mappedJSON) == 0 {
		return s
	}

	var keyValuePairs []string
	var keys []string
	for k := range mappedJSON {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, jsonKey := range keys {
		columnExpr := mappedJSON[jsonKey]
		keyValuePairs = append(keyValuePairs,
			fmt.Sprintf("'%s'", jsonKey),
			columnExpr,
		)
	}

	var formattedColumn string
	if asArrayAggregation {
		orderBy := ""
		if len(orderByClauses) > 0 {
			orderBy = fmt.Sprintf(" ORDER BY %s", orderByClauses[0])
		}
		formattedColumn = fmt.Sprintf("jsonb_agg(DISTINCT jsonb_build_object(%s)%s)", strings.Join(keyValuePairs, ", "), orderBy)
		if condition != "" {
			formattedColumn = fmt.Sprintf("%s FILTER (WHERE %s)", formattedColumn, condition)
		}
	} else {
		formattedColumn = fmt.Sprintf("jsonb_build_object(%s)", strings.Join(keyValuePairs, ", "))
		if condition != "" {
			formattedColumn = fmt.Sprintf("CASE WHEN %s THEN %s ELSE NULL END", condition, formattedColumn)
		}
	}
	formattedColumn = fmt.Sprintf(`%s AS "%s"`, formattedColumn, alias)

	if s.WrapAggregation && !asArrayAggregation {
		s.NestedAggregation = append(s.NestedAggregation, formattedColumn)
	} else {
		s.Columns = append(s.Columns, formattedColumn)
	}

	return s
}

func (s *SelectBuilder) SelectJSONAggregateFunc(alias string, fn func(builder *SelectBuilder)) SQLSelectChainBuilder {
	if alias == "" {
		alias = "json_result"
	}

	s.WrapAggregation = true
	fn(s)

	s.Columns = append(
		s.Columns,
		fmt.Sprintf(`jsonb_build_object(%s) AS "%s"`, strings.Join(s.NestedAggregation, ", "), alias),
	)

	s.WrapAggregation = false
	return s
}
