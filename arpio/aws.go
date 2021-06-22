package arpio

import (
	"fmt"
	"strings"
)

func IsARN(a string) bool {
	return strings.HasPrefix(a, "arn:") && strings.Count(a, ":") >= 5
}

func ValidateArn(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "" {
		return ws, errors
	}

	if !IsARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN", k, value))
		return ws, errors
	}

	return ws, errors
}
