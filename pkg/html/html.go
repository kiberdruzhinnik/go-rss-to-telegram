package html

import (
	"regexp"
)

func ExtractImageLinkFromImgTag(html string) string {
	re := regexp.MustCompile(`<img src="(.*?)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

const regex = `<.*?>`

func SimpleStripAllHTML(s string) string {
	r := regexp.MustCompile(regex)
	return r.ReplaceAllString(s, "")
}
