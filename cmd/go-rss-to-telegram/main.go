package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	gorsstotelegram "github.com/kiberdruzhinnik/go-rss-to-telegram/pkg/go-rss-to-telegram"
)

func readConfiguration() gorsstotelegram.Config {
	ex, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config.yml")

	var cfg gorsstotelegram.Config
	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	if strings.ToLower(cfg.FetchTagsString) == "true" {
		cfg.FetchTags = true
	} else {
		cfg.FetchTags = false
	}

	// split and fix urls
	splittedUrls := strings.Split(cfg.FeedURL, ",")
	feedUrls := make([]string, 0)
	for _, url := range splittedUrls {
		currentUrl := strings.TrimSpace(url)
		if currentUrl == "" {
			continue
		}
		feedUrls = append(feedUrls, currentUrl)
	}
	if len(feedUrls) == 0 {
		log.Fatalln("No feed urls provided")
	}
	cfg.FeedURLsParsed = feedUrls

	telegramChannelID, err := strconv.Atoi(cfg.TelegramChannelIDString)
	if err != nil {
		log.Fatalln(err)
	}
	cfg.TelegramChannelID = int64(telegramChannelID)

	return cfg
}

func main() {
	cfg := readConfiguration()
	gorsstotelegram.Execute(cfg)
}
