package util

import "strings"

func GetFileName(p string) string {
	parts := strings.Split(p, "/")
	nameWithSuffix := parts[len(parts)-1]
	its := strings.Split(nameWithSuffix, ".")
	return its[0]
}
