package mgo

import "unicode"

func ToSnakeCase(str string) string {
	var result string
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result += "_"
			}
			result += string(unicode.ToLower(r))
		} else {
			result += string(r)
		}
	}
	return result
}
