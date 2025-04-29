package notifier

import (
	"fmt"

	"github.com/jamielinux/feed-notifier/internal/config"
	"github.com/mmcdole/gofeed"
)

// Notifier is the interface for sending notifications.
type Notifier interface {
	Notify(feed *config.Feed, item *gofeed.Item) error
}

// NotifierFactory handles the creation of Notifier instances.
type NotifierFactory struct{}

// NewFactory creates a new NotifierFactory.
func NewFactory() *NotifierFactory {
	return &NotifierFactory{}
}

// Create creates notifier instances.
func (f *NotifierFactory) Create(notifierConfig *config.Notifier) (Notifier, error) {
	switch notifierConfig.Type {
	case config.NotifierMattermostWebhook:
		settings := notifierConfig.Settings.(*config.MattermostWebhookSettings)
		return NewMattermostWebhook(settings), nil
	case config.NotifierPushover:
		settings := notifierConfig.Settings.(*config.PushoverSettings)
		return NewPushover(settings), nil
	default:
		return nil, fmt.Errorf("unsupported notifier type: %s", notifierConfig.Type)
	}
}
