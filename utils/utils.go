package utils

import "strings"

func PreReplace(s string) string {
	return s
	if strings.Contains(s, " ") {
		s = "\"" + s + "\""
	}
	return s
}

func IfThen[T any](cond bool, trueVal T, falseVal T) T {
	if cond {
		return trueVal
	} else {
		return falseVal
	}
}
