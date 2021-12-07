package utils

// IsIn checks whether a value is inside the
// given slice of string.
func IsIn(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}
