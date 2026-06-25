package envtools

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// sensitiveKeywords are keywords used to identify sensitive environment variables.
var sensitiveKeywords = []string{
	"PASSWORD", "PASSWD", "SECRET", "TOKEN", "KEY",
	"CREDENTIAL", "AUTH", "PRIVATE", "API_KEY", "APIKEY",
}

// GetEnvVar returns the value of a specific environment variable.
// Practical uses: checking runtime configuration (DB_HOST, PORT, SERVICE_URL).
func GetEnvVar(name string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fmt.Sprintf("Environment variable '%s' does not exist", name)
	}
	if value == "" {
		return fmt.Sprintf("Environment variable '%s' exists but is empty", name)
	}
	// Mask sensitive values
	if isSensitiveKey(name) {
		return fmt.Sprintf("%s = ******* (masked: sensitive key)", name)
	}
	return fmt.Sprintf("%s = %s", name, value)
}

// ListEnvVars lists environment variables by prefix, masking sensitive values.
// Practical uses: checking all environment configuration when debugging a deployment.
func ListEnvVars(prefix string) string {
	all := os.Environ()
	sort.Strings(all)

	prefixUpper := strings.ToUpper(prefix)
	var matches []string
	for _, entry := range all {
		if prefixUpper != "" && !strings.HasPrefix(strings.ToUpper(entry), prefixUpper) {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		if isSensitiveKey(key) {
			matches = append(matches, fmt.Sprintf("  %s=******* (masked)", key))
		} else {
			matches = append(matches, fmt.Sprintf("  %s=%s", key, val))
		}
	}

	if len(matches) == 0 {
		if prefix == "" {
			return "No environment variables found"
		}
		return fmt.Sprintf("No environment variables found with prefix '%s'", prefix)
	}

	label := prefix
	if prefix == "" {
		label = "(all)"
	}
	return fmt.Sprintf(
		"Environment variables (prefix: %s) — %d result(s):\n%s\n%s",
		label, len(matches),
		strings.Repeat("─", 60),
		strings.Join(matches, "\n"),
	)
}

func isSensitiveKey(key string) bool {
	keyUpper := strings.ToUpper(key)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(keyUpper, kw) {
			return true
		}
	}
	return false
}
