package utils

import (
	"strings"
)

func UpperCaseString(value string) string {
	result := strings.ToUpper(value)
	return result
}

func CapitalizeFirstLetter(value string) string {
	capitalized := strings.ToUpper(string(value[0])) + value[1:]
	return capitalized
}
