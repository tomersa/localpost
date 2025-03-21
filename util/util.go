package util

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ReplaceEnvVars replaces placeholders like {VAR} with values from vars.
func ReplaceEnvVars(input string, vars map[string]string) string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		key := strings.Trim(match, "{}")
		if value, ok := vars[key]; ok {
			return value
		}
		return match
	})
}

// ReplaceParams replaces URL parameters with values from vars.
func ReplaceParams(input string, vars map[string]string) (string, error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("error parsing URL %s: %v", input, err)
	}
	q := u.Query()
	for key, values := range q {
		if len(values) > 0 {
			newVal := ReplaceEnvVars(values[0], vars)
			if newVal != values[0] {
				q.Set(key, newVal)
			}
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
