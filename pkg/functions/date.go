package functions

import (
	"fmt"
	"time"
)

var yearConverted = map[string]string{
	"2-digit": "06",
	"numeric": "2006",
}

var monthConverted = map[string]string{
	"2-digit": "01",
	"numeric": "1",
	"short":   "Jan",
}

var dayConverted = map[string]string{
	"2-digit": "02",
	"numeric": "2",
}

var weekdayConverted = map[string]string{
	"short": "Mon",
}

// DateFormat formats a given time.Time object into a string based on the specified locale and date format components.
// Parameters:
//   - t: The time.Time object to format.
//   - locale: A string representing the locale (e.g., "en-us", "en-gb").
//   - weekday: A string key representing the desired weekday format (e.g., "short").
//   - day: A string key representing the desired day format (e.g., "2-digit", "numeric").
//   - month: A string key representing the desired month format (e.g., "2-digit", "numeric", "short").
//   - year: A string key representing the desired year format (e.g., "2-digit", "numeric").
//
// Returns:
//   - A formatted date string based on the specified parameters.
//
// Warning:
//   - Ensure that all inputs (locale, weekday, day, month, year) are properly validated before calling this function.
//     Any invalid or unexpected input may result in incorrect formatting or an empty string being returned.
//
// Usage:
//   - This function is best used when the input data has already been sanitized or pre-converted.
func DateFormat(t time.Time, locale string, weekday string, day string, month string, year string) string {
	switch locale {
	case "en-us":
		if weekday != "" {
			return t.Format(fmt.Sprintf(
				"%s, %s %s, %s", weekdayConverted[weekday], monthConverted[month], dayConverted[day], yearConverted[year]),
			)
		}
		return t.Format(fmt.Sprintf(
			"%s/%s/%s", monthConverted[month], dayConverted[day], yearConverted[year]),
		)
	case "en-gb":
		if month == "short" {
			return t.Format(fmt.Sprintf(
				"%s %s %s", dayConverted[day], monthConverted[month], yearConverted[year]),
			)
		}
		return t.Format(fmt.Sprintf(
			"%s/%s/%s", dayConverted[day], monthConverted[month], yearConverted[year]),
		)
	default:
		return ""
	}
}
