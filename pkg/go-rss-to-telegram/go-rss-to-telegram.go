package gorsstotelegram

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kiberdruzhinnik/go-rss-to-telegram/pkg/kv"
	"github.com/kiberdruzhinnik/go-rss-to-telegram/pkg/utils"
	"github.com/mmcdole/gofeed"
)

const TELEGRAM_MAXIMUM_POST_SIZE = 4096

func selectButtonText(url string) string {
	if strings.Contains(url, "blog.kiberdruzhinnik.ru") {
		return "📖 Читать в блоге"
	} else if strings.Contains(url, "vk.com/video") {
		return "▶️ Смотреть на VK Video"
	} else if strings.Contains(url, "dzen.ru") {
		return "▶️ Смотреть на Dzen"
	} else if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		return "🎥 Смотреть на YouTube"
	} else if strings.Contains(url, "rutube.ru") {
		return "📺 Смотреть на RuTube"
	} else {
		return "👟 Перейти на"
	}
}

func getNewFeedItems(feed *gofeed.Feed, db *kv.KV) []*gofeed.Item {
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

func processNewFeedItems(newFeedItems []*gofeed.Item, tgBot *tgbotapi.BotAPI, db *kv.KV, cfg Config) {
	for _, item := range newFeedItems {
		log.Printf("Processing new entry: %s\n", item.GUID)
		description := utils.SimpleStripAllHTML(item.Description)

		content := ""

		title := strings.TrimSpace(item.Title)
		description = strings.TrimSpace(description)
		link := strings.TrimSpace(item.Link)

		if title != "" {
			content += fmt.Sprintf("<b>%s</b>\n\n", title)
		}

		content += fmt.Sprintf("%s: %s\n\n", selectButtonText(link), link)

		if description != "" {
			content += fmt.Sprintf("%s\n\n", description)
		}

		if len(content) > TELEGRAM_MAXIMUM_POST_SIZE {
			content = utils.EncodeToUTF8(fmt.Sprintf("%s...", content[:TELEGRAM_MAXIMUM_POST_SIZE-3]))
		}
		msg := tgbotapi.NewMessage(cfg.TelegramChannelID, content)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = false

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
				log.Println(err)
				continue
			}
			newFeedItems := getNewFeedItems(feed, db)
			sort.Sort(sort.Reverse(GoFeedItemSlice(newFeedItems)))
			processNewFeedItems(newFeedItems, tgBot, db, cfg)
		}
		log.Printf("Now sleeping for %d minutes\n", cfg.SleepTimeMinutes)
		time.Sleep(time.Minute * time.Duration(cfg.SleepTimeMinutes))
	}
}
