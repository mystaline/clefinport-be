package sql_query

type SQLOperators string

const (
	// ─────────────── Comparison ───────────────

	// Usage: {"age": {Operator: SQLOperatorEqual, Value: 30}}  →  "age" = $1
	SQLOperatorEqual SQLOperators = "="
	// Usage: {"status": {Operator: SQLOperatorNotEqual, Value: "active"}}  →  "status" != $1
	SQLOperatorNotEqual SQLOperators = "!=" // or "<>"
	// Usage: {"score": {Operator: SQLOperatorGreaterThan, Value: 90}}  →  "score" > $1
	SQLOperatorGreaterThan SQLOperators = ">"
	// Usage: {"price": {Operator: SQLOperatorLessThan, Value: 100}}  →  "price" < $1
	SQLOperatorLessThan SQLOperators = "<"
	// Usage: {"age": {Operator: SQLOperatorGTE, Value: 18}}  →  "age" >= $1
	SQLOperatorGTE SQLOperators = ">="
	// Usage: {"age": {Operator: SQLOperatorLTE, Value: 65}}  →  "age" <= $1
	SQLOperatorLTE SQLOperators = "<="

	// ─────────────── Regex ───────────────

	// Usage: {"name": {Operator: SQLOperatorRegexCaseSensitive, Value: "^A"}}  →  "name" ~ $1
	SQLOperatorRegexCaseSensitive SQLOperators = "~"
	// Usage: {"email": {Operator: SQLOperatorRegexCaseInsensitive, Value: "@gmail"}}  →  "email" ~* $1
	SQLOperatorRegexCaseInsensitive SQLOperators = "~*"
	// Usage: {"title": {Operator: SQLOperatorNotRegexCaseSensitive, Value: "test"}}  →  "title" !~ $1
	SQLOperatorNotRegexCaseSensitive SQLOperators = "!~"
	// Usage: {"username": {Operator: SQLOperatorNotRegexCaseInsensitive, Value: "admin"}}  →  "username" !~* $1
	SQLOperatorNotRegexCaseInsensitive SQLOperators = "!~*"

	// ─────────────── Set ───────────────

	// Usage: {"id": {Operator: SQLOperatorIn, Value: []int{1,2,3}}}  →  "id" IN ($1, $2, $3)
	SQLOperatorIn SQLOperators = "IN"
	// Usage: {"id": {Operator: SQLOperatorNotIn, Value: []int{1,2,3}}}  →  "id" NOT IN ($1, $2, $3)
	SQLOperatorNotIn SQLOperators = "NOT IN"

	// ─────────────── Array ───────────────

	// Usage: {"tags": {Operator: SQLOperatorAny, Value: pq.Array([]string{"a","b"})}}
	// →  "tags" = ANY($1)
	SQLOperatorAny SQLOperators = "ANY"

	// ─────────────── Pattern matching ───────────────

	// Usage: {"title": {Operator: SQLOperatorLike, Value: "%hello%"}}  →  "title" LIKE $1
	SQLOperatorLike SQLOperators = "LIKE"
	// Usage: {"title": {Operator: SQLOperatorNotLike, Value: "%test%"}}  →  "title" NOT LIKE $1
	SQLOperatorNotLike SQLOperators = "NOT LIKE"
	// Usage: {"name": {Operator: SQLOperatorILike, Value: "john%"}}  →  "name" ILIKE $1 (case-insensitive, PostgreSQL only)
	SQLOperatorILike SQLOperators = "ILIKE"
	// Usage: {"name": {Operator: SQLOperatorNotILike, Value: "doe%"}}  →  "name" NOT ILIKE $1 (case-insensitive, PostgreSQL only)
	SQLOperatorNotILike SQLOperators = "NOT ILIKE"

	// ─────────────── Null checks ───────────────

	// Usage: {"deleted_at": {Operator: SQLOperatorIsNull}}  →  "deleted_at" IS NULL
	SQLOperatorIsNull SQLOperators = "IS NULL"
	// Usage: {"deleted_at": {Operator: SQLOperatorIsNotNull}}  →  "deleted_at" IS NOT NULL
	SQLOperatorIsNotNull SQLOperators = "IS NOT NULL"

	// ─────────────── Range ───────────────

	// Usage: {"created_at": {Operator: SQLOperatorBetween, Value: []int64{1672531200000, 1675209599000}, IsEpochTime: true}}
	// →  "created_at" BETWEEN to_timestamp($1) AND to_timestamp($2)
	SQLOperatorBetween SQLOperators = "BETWEEN"
	// Usage: {"price": {Operator: SQLOperatorNotBetween, Value: []int{10, 50}}}
	// →  "price" NOT BETWEEN $1 AND $2
	SQLOperatorNotBetween SQLOperators = "NOT BETWEEN"

	// ─────────────── Subqueries / Existence ───────────────

	// Usage: {"": {Operator: SQLOperatorExist, Value: "(SELECT 1 FROM users WHERE active = TRUE)", IsSubQuery: true}}
	// →  EXISTS (SELECT 1 FROM users WHERE active = TRUE)
	SQLOperatorExist SQLOperators = "EXISTS"
	// Usage: {"": {Operator: SQLOperatorNotExist, Value: "(SELECT 1 FROM users WHERE active = TRUE)", IsSubQuery: true}}
	// →  NOT EXISTS (SELECT 1 FROM users WHERE active = TRUE)
	SQLOperatorNotExist SQLOperators = "NOT EXISTS"

	// ─────────────── Raw SQL ───────────────

	// Usage: {"": {Operator: SQLOperatorRaw, Value: `"age" > ? OR "status" = 'active'`, ExtraArgs: []any{30}}}
	// →  age > ? OR status = 'active'
	SQLOperatorRaw SQLOperators = "__RAW__"
)
