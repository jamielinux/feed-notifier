package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/jamielinux/feed-notifier/internal/config"
	"github.com/jamielinux/feed-notifier/internal/logger"
	"github.com/mmcdole/gofeed"
)

// MattermostWebhookNotifier sends notifications via Mattermost webhook.
type MattermostWebhookNotifier struct {
	settings *config.MattermostWebhookSettings
}

// MattermostAttachment represents a Mattermost message attachment.
type MattermostAttachment struct {
	Fallback   string `json:"fallback,omitempty"`
	Color      string `json:"color,omitempty"`
	Pretext    string `json:"pretext,omitempty"`
	Text       string `json:"text,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	AuthorLink string `json:"author_link,omitempty"`
	Title      string `json:"title,omitempty"`
	TitleLink  string `json:"title_link,omitempty"`
	ImageURL   string `json:"image_url,omitempty"`
}

// MattermostMessage represents a Mattermost webhook message.
type MattermostMessage struct {
	Attachments []MattermostAttachment `json:"attachments"`
}

// NewMattermostWebhook creates a new Mattermost webhook notifier.
func NewMattermostWebhook(settings *config.MattermostWebhookSettings) *MattermostWebhookNotifier {
	return &MattermostWebhookNotifier{
		settings: settings,
	}
}

// Notify implements the Notifier interface for MattermostWebhookNotifier.
func (notifier *MattermostWebhookNotifier) Notify(feed *config.Feed, item *gofeed.Item) error {
	logger.Debug("[%s] Sending notification for %s: %s", feed.Notifier, feed.DisplayName, item.Title)

	title := item.Title
	if strings.TrimSpace(title) == "" {
		title = "(no title)"
	}

	if item.PublishedParsed != nil {
		formattedDate := item.PublishedParsed.Format("Jan 2, 2006")
		title = fmt.Sprintf("%s | %s", formattedDate, title)
	}

	text := item.Content
	if strings.TrimSpace(item.Content) == "" {
		text = "(no content)"
	}

	if notifier.settings.HTMLToMarkdown {
		markdown, err := htmltomarkdown.ConvertString(text)
		if err != nil {
			logger.Debug("[%s] Converting HTML to markdown failed for %s: %s", feed.Notifier, feed.DisplayName, item.Title)
		} else {
			text = markdown
		}
	}

	attachment := MattermostAttachment{
		Fallback:   title,
		AuthorName: feed.DisplayName,
		Title:      title,
		TitleLink:  item.Link,
		Text:       text,
	}

	message := MattermostMessage{
		Attachments: []MattermostAttachment{attachment},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to prepare Mattermost notification: %w", err)
	}

	resp, err := http.Post(notifier.settings.Webhook, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send Mattermost webhook notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Mattermost webhook error: %d", resp.StatusCode)
	}

	return nil
}
