package config

import (
	"fmt"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

// Validate ensures that the configuration file is valid.
func (c *Config) Validate() error {
	if err := c.validateFetch(); err != nil {
		return err
	}

	notifierIDs, err := c.validateNotifiers()
	if err != nil {
		return err
	}

	if err := c.validateDefaultNotifier(notifierIDs); err != nil {
		return err
	}

	if err := c.validateFeeds(notifierIDs); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateFetch() error {
	if c.Fetch.Jobs < 0 {
		return fmt.Errorf("fetch.jobs cannot be negative")
	}
	if c.Fetch.Jobs > 10 {
		return fmt.Errorf("fetch.jobs cannot be greater than 10")
	}
	if c.Fetch.Interval <= 0 {
		return fmt.Errorf("fetch.interval cannot be negative")
	}
	return nil
}

func (c *Config) validateNotifiers() (map[string]bool, error) {
	notifierIDs := make(map[string]bool)
	notifierIDs["stdout"] = true

	for i := range c.Notifiers {
		notifier := &c.Notifiers[i]

		if notifier.ID == "" {
			return nil, fmt.Errorf("notifiers must have an id")
		}

		if notifier.ID == "stdout" {
			return nil, fmt.Errorf("notifiers cannot have an id of 'stdout'")
		}

		if _, exists := notifierIDs[notifier.ID]; exists {
			return nil, fmt.Errorf("duplicate notifier id '%s'", notifier.ID)
		}
		notifierIDs[notifier.ID] = true

		if notifier.Type == "" {
			return nil, fmt.Errorf("type must be defined for notifier '%s'", notifier.ID)
		}

		if err := c.validateSettings(notifier); err != nil {
			return nil, err
		}
	}

	return notifierIDs, nil
}

func (c *Config) validateSettings(n *Notifier) error {
	k := koanf.New(".")

	if err := k.Load(confmap.Provider(n.RawSettings, "."), nil); err != nil {
		return fmt.Errorf("failed to load settings for notifier '%s': %v", n.ID, err)
	}

	switch n.Type {
	case NotifierMattermostWebhook:
		var s MattermostWebhookSettings
		if err := k.Unmarshal("", &s); err != nil {
			return fmt.Errorf("invalid settings for notifier '%s': %v", n.ID, err)
		}
		if err := s.Validate(n.ID); err != nil {
			return err
		}
		n.Settings = &s
	case NotifierPushover:
		var s PushoverSettings
		if err := k.Unmarshal("", &s); err != nil {
			return fmt.Errorf("invalid settings for notifier '%s': %v", n.ID, err)
		}
		if err := s.Validate(n.ID); err != nil {
			return err
		}
		n.Settings = &s
	default:
		return fmt.Errorf("type '%s' is invalid for notifier '%s'", n.Type, n.ID)
	}

	return nil
}

func (c *Config) validateDefaultNotifier(notifierIDs map[string]bool) error {
	if _, exists := notifierIDs[c.DefaultNotifier]; !exists {
		return fmt.Errorf("default_notifier '%s' does not match any notifiers", c.DefaultNotifier)
	}
	return nil
}

func (c *Config) validateFeeds(notifierIDs map[string]bool) error {
	feedIDs := make(map[string]bool)

	for i := range c.Feeds {
		feed := &c.Feeds[i]

		if feed.ID == "" {
			return fmt.Errorf("feeds must have an id")
		}

		if feed.DisplayName == "" {
			return fmt.Errorf("display name must be defined for feed '%s'", feed.ID)
		}

		if _, exists := feedIDs[feed.ID]; exists {
			return fmt.Errorf("duplicate feed id '%s'", feed.ID)
		}
		feedIDs[feed.ID] = true

		if feed.URL == "" {
			return fmt.Errorf("url must be defined for feed '%s'", feed.ID)
		}

		if feed.Interval < 0 {
			return fmt.Errorf("interval cannot be negative for feed '%s'", feed.ID)
		}

		if feed.Notifier != "" {
			if _, exists := notifierIDs[feed.Notifier]; !exists {
				return fmt.Errorf("notifier '%s' for feed '%s' does not match any notifiers", feed.Notifier, feed.ID)
			}
		}
	}

	return nil
}
