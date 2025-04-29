package service

import (
	"strconv"
	"strings"

	"github.com/jamielinux/feed-notifier/internal/config"
	"github.com/jamielinux/feed-notifier/internal/db"
	"github.com/jamielinux/feed-notifier/internal/notifier"
	"github.com/mmcdole/gofeed"
)

// getFeedMetadata gets or creates feed metadata.
func (s *Service) getFeedMetadata(feed *config.Feed) (*db.Feed, bool) {
	metadata := s.db.GetFeed(feed.ID)

	firstRun := false
	if metadata == nil {
		firstRun = true
		metadata = &db.Feed{
			FeedID:      feed.ID,
			LastChecked: 0,
		}
	}

	return metadata, firstRun
}

// getArticleID returns a unique identifier for the given article.
func (s *Service) getArticleID(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	if item.Title != "" {
		return item.Title
	}
	return ""
}

// parseMaxAge parses Cache-Control max-age header.
func parseMaxAge(cacheControl string, maximum int64) int64 {
	if cacheControl == "" {
		return 0
	}

	maxAge := int64(0)
	directives := strings.Split(cacheControl, ",")

	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if strings.HasPrefix(directive, "max-age=") {
			parts := strings.Split(directive, "=")
			if len(parts) == 2 {
				val, err := strconv.ParseInt(parts[1], 10, 64)
				if err == nil && val > 0 {
					maxAge = val
					break
				}
			}
			// Only read the first max-age directive
			break
		}
	}

	if maxAge > maximum {
		maxAge = maximum
	}

	return maxAge
}

// shouldFetchFeed determines if we should make a HTTP request for this feed.
func (s *Service) shouldFetchFeed(feed *config.Feed, metadata *db.Feed, now int64) bool {
	if metadata == nil {
		return true
	}

	if metadata.MaxAge > 0 {
		expiryTime := metadata.LastChecked + metadata.MaxAge
		if now < expiryTime {
			return false
		}
	}

	interval := int64(feed.Interval * 60)
	if (now - metadata.LastChecked) < interval {
		return false
	}

	return true
}

// logItems logs items as processed without sending notifications.
func (s *Service) logItems(feed *config.Feed, items []*gofeed.Item) {
	for _, item := range items {
		articleID := s.getArticleID(item)
		if articleID == "" {
			continue
		}
		s.db.LogArticle(feed.ID, articleID)
	}
}

// getNotifierForFeed returns the configured notifier for a feed.
func (s *Service) getNotifierForFeed(feed *config.Feed) notifier.Notifier {
	return s.notifierMap[feed.Notifier]
}
