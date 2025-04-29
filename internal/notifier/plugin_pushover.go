package notifier

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jamielinux/feed-notifier/internal/config"
	"github.com/jamielinux/feed-notifier/internal/logger"
	"github.com/mmcdole/gofeed"
)

// PushoverNotifier sends notifications via Pushover.
type PushoverNotifier struct {
	settings *config.PushoverSettings
}

// NewPushover creates a new Pushover notifier.
func NewPushover(settings *config.PushoverSettings) *PushoverNotifier {
	return &PushoverNotifier{
		settings: settings,
	}
}

// Notify implements the Notifier interface for PushoverNotifier.
func (n *PushoverNotifier) Notify(feed *config.Feed, item *gofeed.Item) error {
	logger.Debug("[%s] Sending notification for %s: %s", feed.Notifier, feed.DisplayName, item.Title)

	title := feed.DisplayName

	message := item.Title
	if strings.TrimSpace(message) == "" {
		message = "(no title)"
	}

	resp, err := http.PostForm("https://api.pushover.net/1/messages.json", url.Values{
		"token":     {n.settings.AppToken},
		"user":      {n.settings.UserKey},
		"title":     {title},
		"url":       {item.Link},
		"url_title": {"Open article"},
		"message":   {message},
	})

	if err != nil {
		return fmt.Errorf("failed to send Pushover notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Pushover API error: %d", resp.StatusCode)
	}

	return nil
}
