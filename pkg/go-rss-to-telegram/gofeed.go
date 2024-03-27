package gorsstotelegram

import "github.com/mmcdole/gofeed"

type GoFeedItemSlice []*gofeed.Item

func (a GoFeedItemSlice) Len() int {
	return len(a)
}

func (a GoFeedItemSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a GoFeedItemSlice) Less(i, j int) bool {
	return a[i].PublishedParsed.After(*a[j].PublishedParsed)
}
