package notifier

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jamielinux/feed-notifier/internal/config"
	"github.com/jamielinux/feed-notifier/internal/logger"
	"github.com/mmcdole/gofeed"
)

// ArticleNotification represents the data to be output as JSON.
type ArticleNotification struct {
	Feed struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
	} `json:"feed"`
	Article struct {
		GUID        string    `json:"guid,omitempty"`
		Title       string    `json:"title"`
		Link        string    `json:"link"`
		Description string    `json:"description,omitempty"`
		Content     string    `json:"content,omitempty"`
		Published   time.Time `json:"published,omitempty"`
		Updated     time.Time `json:"updated,omitempty"`
	} `json:"article"`
	Timestamp time.Time `json:"timestamp"`
}

// StdoutNotifier is a simple notifier that prints to stdout.
type StdoutNotifier struct{}

// NewStdout creates a new stdout notifier.
func NewStdout() *StdoutNotifier {
	return &StdoutNotifier{}
}

// Notify implements the Notifier interface for StdoutNotifier.
func (n *StdoutNotifier) Notify(feed *config.Feed, item *gofeed.Item) error {
	logger.Debug("[stdout] Processing notification for feed '%s', item '%s'",
		feed.DisplayName, item.Title)

	notification := ArticleNotification{
		Timestamp: time.Now(),
	}

	notification.Feed.ID = feed.ID
	notification.Feed.DisplayName = feed.DisplayName
	notification.Feed.URL = feed.URL

	notification.Article.Title = item.Title
	notification.Article.Link = item.Link
	notification.Article.GUID = item.GUID

	if item.Description != "" {
		notification.Article.Description = item.Description
	}

	if item.Content != "" {
		notification.Article.Content = item.Content
	}

	if item.PublishedParsed != nil {
		notification.Article.Published = *item.PublishedParsed
	}

	if item.UpdatedParsed != nil {
		notification.Article.Updated = *item.UpdatedParsed
	}

	jsonData, err := json.MarshalIndent(notification, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}

	log.Println(string(jsonData))
	return nil
}
