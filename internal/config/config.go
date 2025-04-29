package config

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Feed represents an RSS/Atom feed to be monitored.
type Feed struct {
	ID          string `koanf:"id"`
	URL         string `koanf:"url"`
	DisplayName string `koanf:"display_name"`
	Interval    int    `koanf:"interval"`
	Notifier    string `koanf:"notifier"`
}

const (
	NotifierMattermostWebhook = "mattermost_webhook"
	NotifierPushover          = "pushover"
	NotifierStdout            = "stdout"
)

// NotifierSettings is an interface that all notifier settings must implement.
type NotifierSettings interface {
	Validate(notifierID string) error
}

// MattermostWebhookSettings contains options for Mattermost Webhook notifications.
type MattermostWebhookSettings struct {
	Webhook        string `koanf:"webhook"`
	HTMLToMarkdown bool   `koanf:"html_to_markdown"`
}

// Validate implements the NotifierSettings interface for MattermostWebhookSettings.
func (s *MattermostWebhookSettings) Validate(notifierID string) error {
	if s.Webhook == "" {
		return fmt.Errorf("settings.webhook must be defined for notifier '%s'", notifierID)
	}
	return nil
}

// PushoverSettings contains options for Pushover notifications.
type PushoverSettings struct {
	AppToken string `koanf:"app_token"`
	UserKey  string `koanf:"user_key"`
}

// Validate implements the NotifierSettings interface for PushoverSettings.
func (s *PushoverSettings) Validate(notifierID string) error {
	if s.AppToken == "" {
		return fmt.Errorf("settings.app_token must be defined for notifier '%s'", notifierID)
	}
	if s.UserKey == "" {
		return fmt.Errorf("settings.user_key must be defined for notifier '%s'", notifierID)
	}
	return nil
}

// Notifier is a method of sending notifications.
type Notifier struct {
	ID          string                 `koanf:"id"`
	Type        string                 `koanf:"type"`
	RawSettings map[string]interface{} `koanf:"settings"`
	Settings    NotifierSettings       `koanf:"-"`
}

// Config represents the complete configuration for the program.
type Config struct {
	Database string `koanf:"database"`
	Debug    bool   `koanf:"debug"`
	Fetch    struct {
		Jobs     int `koanf:"jobs"`
		Interval int `koanf:"interval"`
	} `koanf:"fetch"`
	Notifiers       []Notifier `koanf:"notifiers"`
	DefaultNotifier string     `koanf:"default_notifier"`
	Feeds           []Feed     `koanf:"feeds"`
}

// Load loads the config file and creates a new Config.
func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	config.setDefaults()

	return &config, nil
}

// setDefaults sets default settings if unspecified.
func (c *Config) setDefaults() {
	if c.Database == "" {
		c.Database = os.ExpandEnv("$HOME/.local/share/feed-notifier/db")
	} else {
		c.Database = os.ExpandEnv(c.Database)
	}

	if c.Fetch.Jobs == 0 {
		c.Fetch.Jobs = 3
	}

	if c.Fetch.Interval == 0 {
		c.Fetch.Interval = 60
	}

	if c.DefaultNotifier == "" {
		c.DefaultNotifier = "stdout"
	}

	for i := range c.Feeds {
		feed := &c.Feeds[i]
		if feed.Interval == 0 {
			feed.Interval = c.Fetch.Interval
		}
		if feed.Notifier == "" {
			feed.Notifier = c.DefaultNotifier
		}
	}
}
