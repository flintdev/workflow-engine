package util

import (
	"encoding/json"
	"fmt"
)

func ConvertMapToJsonString(m map[string]string) string {
	empData, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonStr := string(empData)
	return jsonStr
}

func ConvertJsonStringToMap(s string) map[string]string {
	m := make(map[string]string)

	err := json.Unmarshal([]byte(s), &m)

	if err != nil {
		fmt.Println(err.Error())
	}
	return m
}
