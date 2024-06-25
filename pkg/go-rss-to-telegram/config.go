package gorsstotelegram

type Config struct {
	// comma separated list of FEEDS
	// example:
	// https://link.to/rss.xml,https://another.website/atom.xml
	FeedURL                 string `yaml:"FEED_URL" env:"FEED_URL"`
	FeedURLsParsed          []string
	DatabasePath            string `yaml:"DATABASE_PATH" env:"DATABASE_PATH"`
	SleepTimeMinutes        int    `yaml:"SLEEP_TIME_MINUTES" env:"SLEEP_TIME_MINUTES"`
	TelegramBotToken        string `yaml:"TELEGRAM_BOT_TOKEN" env:"TELEGRAM_BOT_TOKEN"`
	TelegramChannelIDString string `yaml:"TELEGRAM_CHANNEL_ID" env:"TELEGRAM_CHANNEL_ID"`
	TelegramChannelID       int64
	FetchTagsString         string `yaml:"FETCH_TAGS" env:"FETCH_TAGS"`
	FetchTags               bool
}
