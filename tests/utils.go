package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

const UUIDRegexMatch = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"

func UnmarshalJSON(data []byte) (map[string]any, error) {
	var JSON map[string]any
	err := json.Unmarshal(data, &JSON)
	return JSON, err
}

func GetBody(response *http.Response) []byte {
	if response.ContentLength == 0 {
		return nil
	}

	bodyData := make([]byte, response.ContentLength)
	_, _ = response.Body.Read(bodyData)
	return bodyData
}

func GetBodyJSON(response *http.Response) (map[string]any, error) {
	return UnmarshalJSON(GetBody(response))
}

func JSONMatches(expected string, actual string) bool {
	expectedJSON, _ := UnmarshalJSON([]byte(expected))
	actualJSON, _ := UnmarshalJSON([]byte(actual))
	return JSONMapMatches(expectedJSON, actualJSON)
}

func JSONMapMatches(expected map[string]any, actual map[string]any) bool {
	if len(expected) != len(actual) {
		return false
	}

	for key, expected_value := range expected {
		actual_value, exists := actual[key]
		if !exists {
			return false
		}

		switch v := expected_value.(type) {
		case []interface{}:
			for _, expected_inner_value := range v {
				match := false
				for _, actual_inner_value := range actual_value.([]interface{}) {
					if JSONMapMatches(expected_inner_value.(map[string]any), actual_inner_value.(map[string]any)) {
						match = true
						break
					}
				}

				if !match {
					return false
				}
			}
		case string:
			match, _ := regexp.MatchString(v, actual_value.(string))
			if !match {
				return false
			}
		default:
			log.Fatalf("Un-expected type: %T", v)
		}
	}

	return true
}

func errorJSON(statusCode int, message string, details string) string {
	return fmt.Sprintf(`{
		"error": "%s",
		"message": "%s",
		"details": %s
	}`, http.StatusText(statusCode), message, details)
}

func ErrorJSON(statusCode int, message string, details string) string {
	return errorJSON(statusCode, message, fmt.Sprintf(`"%s"`, details))
}

func ValidationErrorJSON(message string, field string, reason string, value string) string {
	return errorJSON(http.StatusUnprocessableEntity, message, fmt.Sprintf(`[{
		"field": "%s",
		"reason": "%s",
		"value": "%s"
	}]`, field, reason, value))
}
