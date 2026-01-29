package logging

// FilterSensitiveFields removes sensitive data from logs
func FilterSensitiveFields(body map[string]interface{}) map[string]interface{} {
	sensitiveFields := []string{"password", "token", "secret", "api_key", "apikey"}

	filtered := make(map[string]interface{})
	for key, value := range body {
		// Check if field is sensitive
		isSensitive := false
		for _, sensitive := range sensitiveFields {
			if key == sensitive {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			filtered[key] = "***REDACTED***"
		} else {
			filtered[key] = value
		}
	}

	return filtered
}
