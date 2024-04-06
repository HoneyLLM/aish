package utils

import (
	"io"
	"os"
	"strings"
)

func GetWriter(path string) (io.Writer, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func GetLastLine(s string) string {
	lastIndex := strings.LastIndex(s, "\n")
	if lastIndex == -1 {
		return s
	}
	return s[lastIndex+1:]
}

func RemoveLastLine(s string) string {
	lastIndex := strings.LastIndex(s, "\n")
	if lastIndex == -1 {
		return ""
	}
	return s[:lastIndex+1]
}
