package util

import (
	"encoding/json"
)

func ConvertMapToJsonString(m map[string]string) (string, error) {
	empData, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	jsonStr := string(empData)
	return jsonStr, nil
}

func ConvertJsonStringToStringMap(s string) (map[string]string, error) {
	m := make(map[string]string)

	err := json.Unmarshal([]byte(s), &m)

	if err != nil {
		return m, nil
	}
	return m, nil
}

func ConvertJsonStringToMap(s string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	err := json.Unmarshal([]byte(s), &m)

	if err != nil {
		return m, nil
	}
	return m, nil
}
