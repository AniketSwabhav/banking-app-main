package util

import "strings"

// IsEmpty Check str is Empty Or Not.
func IsEmpty(str string) bool {
	if len(strings.TrimSpace(str)) == 0 {
		return true
	}
	return false
}
