package llm

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	// ErrTimeout is returned when the LLM request times out
	ErrTimeout = errors.New("llm request timeout")

	// ErrInvalidResponse is returned when the LLM response is invalid
	ErrInvalidResponse = errors.New("invalid llm response")

	// ErrAPIError is returned when the LLM API returns an error
	ErrAPIError = errors.New("llm api error")
)

// ParseExtractedData parses the extracted data JSON string into a map
func ParseExtractedData(jsonData string) (map[string]interface{}, error) {
	if jsonData == "" {
		return nil, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to parse extracted data: %w", err)
	}

	return data, nil
}

// MarshalExtractedData converts extracted data map to JSON string
func MarshalExtractedData(data map[string]interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal extracted data: %w", err)
	}

	return string(jsonBytes), nil
}

// GetStringSlice safely gets a string slice from extracted data
func GetStringSlice(data map[string]interface{}, key string) []string {
	if data == nil {
		return nil
	}

	value, ok := data[key]
	if !ok {
		return nil
	}

	// Handle []interface{} from JSON unmarshaling
	if arr, ok := value.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, v := range arr {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}

	// Handle []string directly
	if arr, ok := value.([]string); ok {
		return arr
	}

	return nil
}
