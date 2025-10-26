package functions

import (
	"encoding/json"
	"log"
)

// FormattedLog prints any Go object or array as a beautified JavaScript constant.
func FormattedLog(obj interface{}) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// Marshal with indentation for readability
	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Printf("%v", err.Error())
	}

	log.Printf("%v", string(jsonBytes))
}
