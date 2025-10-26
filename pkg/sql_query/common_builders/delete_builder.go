package common_builders

import (
	"log"

	"github.com/mystaline/clefinport-be/pkg/sql_query"
)

func DeleteBuilder(tableName string, query map[string]sql_query.SQLCondition) (string, []interface{}) {
	res, args, err := sql_query.NewSQLDeleteBuilder(tableName).
		Delete("id").
		Where(query).
		Build()

	if err != nil {
		log.Println(err)
	}

	return res, args
}
