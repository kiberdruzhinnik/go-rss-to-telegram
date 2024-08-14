package utils

import (
	"regexp"
	"unicode/utf8"
)

func EncodeToUTF8(input string) string {
	encoded := make([]byte, 0, len(input))
	for _, r := range input {
		b := make([]byte, 4)
		n := utf8.EncodeRune(b, r)
		encoded = append(encoded, b[:n]...)
	}
	return string(encoded)
}

const STRIP_HTML_REGEX = `<.*?>`

func SimpleStripAllHTML(s string) string {
	r := regexp.MustCompile(STRIP_HTML_REGEX)
	return r.ReplaceAllString(s, "")
}
