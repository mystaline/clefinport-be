package sql_query

import (
	"fmt"
)

func GenerateCTEOption(
	sourceBuilder *SelectBuilder,
	mainBuilder *SelectBuilder,
	aliasName string,
	refTable string,
	optionsQuery bool,
	labelKey string,
	labelValue string,
	customSort ...Sort,
) {
	if optionsQuery {
		labelKey = escapeQuoteColumns(labelKey)
		labelValue = escapeQuoteColumns(labelValue)
		snakeAliasName := CamelToSnake(aliasName)

		cteBuilder := sourceBuilder.
			ClearSelects().
			Distinct(
				fmt.Sprintf(`%s AS "value"`, labelValue),
				labelValue,
			).
			Select(
				fmt.Sprintf(`%s AS "key"`, labelKey),
				fmt.Sprintf(`"%s"."id" AS "id"`, refTable),
			)

		mainBuilder.WithCTEBuilder(snakeAliasName, cteBuilder.(*SelectBuilder).SQLEloquentQuery)
		mainBuilder.LeftJoin(
			snakeAliasName,
			fmt.Sprintf(`"%s"."id" = "%s"."id"`, snakeAliasName, refTable),
		)

		orderString := fmt.Sprintf(`"%s"."key" ASC`, snakeAliasName)
		if len(customSort) > 0 {
			orderString = fmt.Sprintf(`%s %s`, customSort[0].SortBy, "ASC")
		}

		// Wrap the original jsonb_agg with COALESCE
		mainBuilder.Select(
			fmt.Sprintf(
				`COALESCE(
					jsonb_agg(
						jsonb_build_object(
							'label', "%s"."key",
							'value', "%s"."value"
						)
						ORDER BY %s
					) FILTER (WHERE "%s"."value" IS NOT NULL AND "%s"."value" != '' AND "%s"."key" != ''),
					'[]'::jsonb
				) AS "%s"`, // The default value '[]'::jsonb is added
				snakeAliasName,
				snakeAliasName,
				orderString,
				snakeAliasName,
				snakeAliasName,
				snakeAliasName,
				aliasName,
			),
		)

	} else {
		mainBuilder.Select(
			fmt.Sprintf(
				`'[]'::jsonb AS "%s"`,
				aliasName,
			),
		)
	}
}
