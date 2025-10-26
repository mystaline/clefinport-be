package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mystaline/clefinport-be/pkg/functions"
)

// QueryTransform is a type alias for string, representing the type of transformation to be applied to a query value.
type QueryTransform string

// Constants defining the supported query transformations.
const (
	TransformString  QueryTransform = "string"  // Transform value to string
	TransformInt     QueryTransform = "int"     // Transform value to int
	TransformFloat32 QueryTransform = "float32" // Transform value to float32
	TransformFloat64 QueryTransform = "float64" // Transform value to float64
	TransformBool    QueryTransform = "bool"    // Transform value to bool
	TransformArray   QueryTransform = "array"   // Transform value to array
	TransformDate    QueryTransform = "date"    // Transform value to time.Time
)

// queryTransformMap maps string representations of transformations to their corresponding QueryTransform constants.
var queryTransformMap = map[string]QueryTransform{
	"string":  TransformString,
	"int":     TransformInt,
	"float32": TransformFloat32,
	"float64": TransformFloat64,
	"bool":    TransformBool,
	"array":   TransformArray,
	"date":    TransformDate,
}

// getQueryTransform returns the QueryTransform constant corresponding to the given transform string.
// If the transform string is invalid, it returns an error.
func getQueryTransform(transform string) (QueryTransform, error) {
	transformType, exist := queryTransformMap[transform]
	if !exist {
		return "", fmt.Errorf("invalid query transform: %s", transform)
	}
	return transformType, nil
}

// transformQueryValue transforms the given string value based on the specified transform type.
// It returns the transformed value as an interface{}.
func transformQueryValue(value string, transform string) interface{} {
	transformType, err := getQueryTransform(transform)
	if err != nil {
		fmt.Println("Error getting query transform:", err)
		return value
	}

	switch transformType {
	case TransformInt:
		v, _ := strconv.Atoi(value)
		return v
	case TransformFloat32:
		v, _ := strconv.ParseFloat(value, 32)
		return v
	case TransformFloat64:
		v, _ := strconv.ParseFloat(value, 64)
		return v
	case TransformBool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil // if value isn't a valid boolean string, return nil
		}
		return v
	case TransformArray:
		var v []interface{}
		json.Unmarshal([]byte(value), &v)
		return v
	case TransformDate:
		v, err := strconv.Atoi(value)
		if err != nil {
			v = 0
		}

		seconds := int64(v / 1000)             // Convert milliseconds to seconds
		nanoseconds := int64((v % 1000) * 1e6) // Convert remainder to nanoseconds
		t := time.Unix(seconds, nanoseconds)

		return t
	default:
		return value
	}
}

// ParseQuery parses a map of string key-value pairs into a struct of type T.
// It uses the struct's field tags to determine how to transform each value.
// Returns a pointer to the parsed struct and an error if any occurs.
func ParseQuery[T any](query map[string]string) (*T, error) {
	t := reflect.TypeFor[T]()

	parsedQuery := map[string]interface{}{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		rawKey := field.Tag.Get("json")
		key := strings.ReplaceAll(rawKey, ",omitempty", "")
		transform := field.Tag.Get("transform")
		transformValue := transformQueryValue(query[key], transform)
		isZeroValue := transformValue == nil
		if transformValue != nil {
			isZeroValue = functions.IsZeroValue(transformValue)
		}
		if !strings.Contains(rawKey, "omitempty") || !isZeroValue {
			parsedQuery[key] = transformValue
		}
	}

	jsonData, err := json.Marshal(parsedQuery)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	var parsedData T
	err = json.Unmarshal(jsonData, &parsedData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil, err
	}

	return &parsedData, nil
}
