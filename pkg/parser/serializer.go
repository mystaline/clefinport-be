package parser

import "encoding/json"

// SerializeJSON serializes the provided json data into struct
func SerializeJSON[T any](data any, out *T) error {
	marshalledData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(marshalledData, out)
	if err != nil {
		return err
	}

	return nil
}
