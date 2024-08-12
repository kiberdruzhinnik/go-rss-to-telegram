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

const TELEGRAM_MAXIMUM_POST_SIZE = 4096
const TELEGRAM_MAXIMUM_PHOTO_SIZE = 1024

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

		content := ""

		title := strings.TrimSpace(item.Title)
		image = strings.TrimSpace(image)
		description = strings.TrimSpace(description)
		link := strings.TrimSpace(item.Link)

		if title != "" {
			content += fmt.Sprintf("%s\n", title)
		}
		if description != "" {
			content += fmt.Sprintf("%s\n\n", description)
		}
		if link != "" {
			content += fmt.Sprintf("%s\n\n", link)
		}

		log.Println(content)

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
			if len(content) > TELEGRAM_MAXIMUM_PHOTO_SIZE {
				content = fmt.Sprintf("%s...", content[:TELEGRAM_MAXIMUM_PHOTO_SIZE-3])
			}
			msg.Caption = content
			message = msg
		} else {
			if len(content) > TELEGRAM_MAXIMUM_POST_SIZE {
				content = fmt.Sprintf("%s...", content[:TELEGRAM_MAXIMUM_POST_SIZE-3])
			}
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
		for _, feedUrl := range cfg.FeedURLsParsed {
			log.Printf("Now checking feed %s\n", feedUrl)
			feedParser := gofeed.NewParser()
			feed, err := feedParser.ParseURL(feedUrl)
			if err != nil {
				log.Fatalln(err)
			}
			newFeedItems := GetNewFeedItems(feed, db)
			sort.Sort(sort.Reverse(GoFeedItemSlice(newFeedItems)))
			ProcessNewFeedItems(newFeedItems, tgBot, db, cfg)
		}
		log.Printf("Now sleeping for %d minutes\n", cfg.SleepTimeMinutes)
		time.Sleep(time.Minute * time.Duration(cfg.SleepTimeMinutes))
	}

}
