package gorsstotelegram

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	khtml "github.com/kiberdruzhinnik/go-rss-to-telegram/pkg/html"
	"github.com/kiberdruzhinnik/go-rss-to-telegram/pkg/kv"
	"github.com/mmcdole/gofeed"
)

func ExtractTags(link string) ([]string, error) {
	response, err := http.Get(link)
	if err != nil {
		return []string{}, err
	}
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return []string{}, err
	}
	return khtml.ExtractTagsFromMeta(string(bytes))
}

func GetNewFeedItems(feed *gofeed.Feed, db *kv.KV) []*gofeed.Item {
	var newFeedItems []*gofeed.Item
	for _, item := range feed.Items {
		exists, err := db.Exists(item.GUID)
		if err != nil {
			continue
		}
		if exists {
			continue
		}
		newFeedItems = append(newFeedItems, item)
	}
	return newFeedItems
}

func ProcessNewFeedItems(newFeedItems []*gofeed.Item, tgBot *tgbotapi.BotAPI, db *kv.KV, cfg Config) {
	for _, item := range newFeedItems {
		log.Printf("Processing new entry: %s\n", item.GUID)
		image := khtml.ExtractImageLinkFromImgTag(item.Description)
		description := khtml.SimpleStripAllHTML(item.Description)

		fetchedTags := []string{}
		if cfg.FetchTags {
			tags, err := ExtractTags(item.Link)
			if err != nil {
				log.Println(err)
				continue
			}
			fetchedTags = tags
		}

		content := fmt.Sprintf("%s\n%s\n\n%s\n\n%s", item.Title, description, item.Link, strings.Join(fetchedTags, " "))
		content = strings.TrimSpace(content)

		var message tgbotapi.Chattable

		if image != "" {
			response, err := http.Get(image)
			if err != nil {
				log.Println(err)
				continue
			}
			bytes, err := io.ReadAll(response.Body)
			if err != nil {
				log.Println(err)
				response.Body.Close()
				continue
			}
			response.Body.Close()
			imagePart := tgbotapi.FileBytes{Name: item.Title, Bytes: bytes}
			msg := tgbotapi.NewPhotoUpload(cfg.TelegramChannelID, imagePart)
			msg.Caption = content
			message = msg
		} else {
			msg := tgbotapi.NewMessage(cfg.TelegramChannelID, content)
			message = msg
		}

		_, err := tgBot.Send(message)
		if err != nil {
			log.Println(err)
			continue
		}

		db.Set(item.GUID, "1")
	}
}

func Execute(cfg Config) {
	log.Printf("Initializing database at %s\n", cfg.DatabasePath)
	db, err := kv.NewBadgerDb(cfg.DatabasePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	tgBot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		log.Printf("Now checking feed %s\n", cfg.FeedURL)
		feedParser := gofeed.NewParser()
		feed, err := feedParser.ParseURL(cfg.FeedURL)
		if err != nil {
			log.Fatalln(err)
		}

		newFeedItems := GetNewFeedItems(feed, db)
		sort.Sort(sort.Reverse(GoFeedItemSlice(newFeedItems)))
		ProcessNewFeedItems(newFeedItems, tgBot, db, cfg)

		log.Printf("Now sleeping for %d minutes\n", cfg.SleepTimeMinutes)
		time.Sleep(time.Minute * time.Duration(cfg.SleepTimeMinutes))
	}

}
