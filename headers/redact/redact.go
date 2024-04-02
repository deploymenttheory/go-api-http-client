// headers/redact/redact.go
package redact

// RedactSensitiveHeaderData redacts sensitive data based on the hideSensitiveData flag.
func RedactSensitiveHeaderData(hideSensitiveData bool, key, value string) string {
	if hideSensitiveData {
		// Define sensitive data keys that should be redacted.
		sensitiveKeys := map[string]bool{
			"AccessToken":   true,
			"Authorization": true,
		}

		if _, found := sensitiveKeys[key]; found {
			return "REDACTED"
		}
	}
	return value
}
