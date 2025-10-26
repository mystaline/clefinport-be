package functions

import (
	"encoding/json"
	"fmt"
)

func JSONStringify(val any) (string, error) {
	jsonStr, err := json.Marshal(val)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	return string(jsonStr), nil
}

func JSONParse[T any](stringified string) (T, error) {
	var parsed T
	err := json.Unmarshal([]byte(stringified), &parsed)
	if err != nil {
		return parsed, fmt.Errorf("%v", err)
	}

	return parsed, nil
}
