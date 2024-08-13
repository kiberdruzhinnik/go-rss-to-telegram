package gorsstotelegram

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	khtml "github.com/kiberdruzhinnik/go-rss-to-telegram/pkg/html"
	"github.com/kiberdruzhinnik/go-rss-to-telegram/pkg/kv"
	"github.com/mmcdole/gofeed"
)

const TELEGRAM_MAXIMUM_POST_SIZE = 4096
const TELEGRAM_MAXIMUM_PHOTO_SIZE = 1024

func SelectButtonText(url string) string {
	if strings.Contains(url, "blog.kiberdruzhinnik.ru") {
		return "📖 Читать в блоге"
	} else if strings.Contains(url, "vk.com/video") {
		return "📺 Смотреть на VK Video"
	} else if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		return "📺 Смотреть на YouTube"
	} else if strings.Contains(url, "rutube.ru") {
		return "📺 Смотреть на RuTube"
	} else {
		return "🚶 Перейти на"
	}
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
		// image := khtml.ExtractImageLinkFromImgTag(item.Description)
		description := khtml.SimpleStripAllHTML(item.Description)

		content := ""

		title := strings.TrimSpace(item.Title)
		// image = strings.TrimSpace(image)
		description = strings.TrimSpace(description)
		link := strings.TrimSpace(item.Link)

		if title != "" {
			content += fmt.Sprintf("⚠️⚠️⚠️<b>%s</b>⚠️⚠️⚠️\n\n", title)
		}

		content += fmt.Sprintf("%s: %s\n\n", SelectButtonText(link), link)

		if description != "" {
			content += fmt.Sprintf("%s\n\n", description)
		}

		if len(content) > TELEGRAM_MAXIMUM_POST_SIZE {
			content = content[:TELEGRAM_MAXIMUM_POST_SIZE]
		}

		if len(content) > TELEGRAM_MAXIMUM_POST_SIZE {
			content = encodeToUTF8(fmt.Sprintf("%s...", content[:TELEGRAM_MAXIMUM_POST_SIZE-3]))
		}
		msg := tgbotapi.NewMessage(cfg.TelegramChannelID, content)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = false
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(SelectButtonText(link), link),
			),
		)

		_, err := tgBot.Send(msg)
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

func encodeToUTF8(input string) string {
	encoded := make([]byte, 0, len(input))
	for _, r := range input {
		b := make([]byte, 4)
		n := utf8.EncodeRune(b, r)
		encoded = append(encoded, b[:n]...)
	}
	return string(encoded)
}
