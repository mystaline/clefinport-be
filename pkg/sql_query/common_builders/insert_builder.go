package common_builders

import (
	"log"

	"github.com/mystaline/clefinport-be/pkg/sql_query"
)

func InsertBuilder(tableName string, body interface{}, returningColumn ...string) (string, []interface{}) {
	res, args, err := sql_query.NewSQLInsertBuilder(tableName).
		Insert(body, returningColumn...).
		Build()
	if err != nil {
		log.Println(err)
	}

	return res, args
}
