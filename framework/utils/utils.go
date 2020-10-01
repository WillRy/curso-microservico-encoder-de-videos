package utils

import (
	"encoding/json"
)

//IsJson -> validate if is JSON
func IsJson(s string) error {
	var js struct{}

	if err := json.Unmarshal([]byte(s), &js); err != nil {
		return err
	}
	return nil
}
