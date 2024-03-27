package html

import (
	"fmt"
	"regexp"
	"strings"
)

func ExtractImageLinkFromImgTag(html string) string {
	re := regexp.MustCompile(`<img src="(.*?)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func ExtractTagsFromMeta(html string) ([]string, error) {
	re := regexp.MustCompile(`<meta name=keywords content="(.*?)">`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return []string{}, fmt.Errorf("no tags found")
	}

	splits := strings.Split(matches[1], ",")
	out := make([]string, len(splits))
	for i, s := range splits {
		out[i] = fmt.Sprintf("#%s", strings.TrimSpace(s))
	}

	return out, nil
}

const regex = `<.*?>`

func SimpleStripAllHTML(s string) string {
	r := regexp.MustCompile(regex)
	return r.ReplaceAllString(s, "")
}
