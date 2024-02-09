// logger.go
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

// LogEntry represents the structure of a log entry.
type LogEntry struct {
	Level      string                 `json:"level"`
	Timestamp  string                 `json:"timestamp"`
	Message    string                 `json:"msg"`
	Additional map[string]interface{} `json:"-"`
}

// UnmarshalJSONLog customizes the unmarshalling of LogEntry to handle arbitrary key-value pairs.
func (le *LogEntry) UnmarshalJSONLog(data []byte) error {
	// Unmarshal fixed fields first
	type Alias LogEntry
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(le),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal additional fields into a map
	var additionalFields map[string]interface{}
	if err := json.Unmarshal(data, &additionalFields); err != nil {
		return err
	}

	// Remove fixed fields from the map
	delete(additionalFields, "level")
	delete(additionalFields, "timestamp")
	delete(additionalFields, "msg")

	le.Additional = additionalFields
	return nil
}

// ProcessLogs parses and outputs the log entries in a human-readable form.
func ProcessLogs(logEntries []string) {
	for _, logEntryStr := range logEntries {
		var logEntry LogEntry
		if err := json.Unmarshal([]byte(logEntryStr), &logEntry); err != nil {
			log.Printf("Failed to parse log entry: %v", err)
			continue
		}

		fmt.Printf("[%s] %s: %s\n", logEntry.Timestamp, logEntry.Level, logEntry.Message)
		for key, value := range logEntry.Additional {
			valueStr := formatValue(value)
			fmt.Printf("  %s: %s\n", key, valueStr)
		}
		fmt.Println()
	}
}

// formatValue returns a string representation of the value, considering its type.
func formatValue(value interface{}) string {
	if value == nil {
		return "null"
	}
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", value)
	case reflect.Int, reflect.Int64, reflect.Float64:
		return fmt.Sprintf("%v", value)
	case reflect.Bool:
		return fmt.Sprintf("%t", value)
	case reflect.Map, reflect.Slice, reflect.Array:
		bytes, err := json.Marshal(value)
		if err != nil {
			return "error"
		}
		return string(bytes)
	default:
		return fmt.Sprintf("%v", value)
	}
}
