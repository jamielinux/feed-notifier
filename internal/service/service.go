package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jamielinux/feed-notifier/internal/config"
	"github.com/jamielinux/feed-notifier/internal/db"
	"github.com/jamielinux/feed-notifier/internal/logger"
	"github.com/jamielinux/feed-notifier/internal/notifier"
	"github.com/mmcdole/gofeed"
)

// Service manages fetching feeds and sending notifications.
type Service struct {
	config      *config.Config
	db          *db.DB
	httpClient  *http.Client
	parser      *gofeed.Parser
	notifierMap map[string]notifier.Notifier

	// concurrency
	ctx       context.Context
	cancel    context.CancelFunc
	ticker    *time.Ticker
	semaphore chan struct{}
	wg        sync.WaitGroup
}

// New creates a Service instance.
func New(config *config.Config, database *db.DB) (*Service, error) {
	ctx, cancel := context.WithCancel(context.Background())
	semaphore := make(chan struct{}, config.Fetch.Jobs)

	service := &Service{
		config:      config,
		db:          database,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		parser:      gofeed.NewParser(),
		notifierMap: make(map[string]notifier.Notifier),

		// concurrency
		ctx:       ctx,
		cancel:    cancel,
		ticker:    time.NewTicker(1 * time.Minute),
		semaphore: semaphore,
	}

	if err := service.initNotifiers(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize notifiers: %w", err)
	}

	return service, nil
}

// initNotifiers sets up the notifiers.
func (s *Service) initNotifiers() error {
	// Add the default built-in stdout notifier.
	s.notifierMap["stdout"] = notifier.NewStdout()

	// Add notifiers from config file.
	factory := notifier.NewFactory()
	for _, n := range s.config.Notifiers {
		notifierInstance, err := factory.Create(&n)
		if err != nil {
			return fmt.Errorf("failed to create notifier '%s': %w", n.ID, err)
		}
		s.notifierMap[n.ID] = notifierInstance
	}

	return nil
}

// Start starts the service.
func (s *Service) Start() {
	logger.Debug("Starting service...")
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.processAllFeeds()
		for {
			select {
			case <-s.ticker.C:
				s.processAllFeeds()
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the service.
func (s *Service) Stop() {
	logger.Debug("Stopping service...")
	s.ticker.Stop()
	s.cancel()
	s.wg.Wait()
}

// processAllFeeds checks all feeds and fetches any that are due to be fetched.
func (s *Service) processAllFeeds() {
	var wg sync.WaitGroup

	for _, feed := range s.config.Feeds {
		select {
		case <-s.ctx.Done():
			// context was cancelled
			return
		default:
			// continue
		}

		feedCopy := feed
		metadata := s.db.GetFeed(feed.ID)
		now := time.Now().Unix()
		if !s.shouldFetchFeed(&feedCopy, metadata, now) {
			continue
		}

		select {
		case s.semaphore <- struct{}{}:
			// got a slot, continue processing
		case <-s.ctx.Done():
			// context was cancelled while waiting for a slot
			return
		}

		wg.Add(1)
		go func(f config.Feed) {
			defer wg.Done()
			defer func() { <-s.semaphore }() // make sure to release the slot
			if err := s.processFeed(&f); err != nil {
				log.Printf("Error processing feed '%s': %v", f.ID, err)
			}
		}(feedCopy)
	}

	wg.Wait()
	logger.Debug("Finished processing feeds")
}

// processFeed handles fetching and processing a single feed.
func (s *Service) processFeed(feed *config.Feed) error {
	logger.Debug("Processing feed: %s (%s)", feed.ID, feed.URL)

	metadata, firstRun := s.getFeedMetadata(feed)

	parsedFeed, httpStatus, err := s.fetchFeed(feed, metadata)
	if err != nil {
		return fmt.Errorf("fetch error: %w", err)
	}

	metadata.LastChecked = time.Now().Unix()
	s.db.UpdateFeed(metadata)

	if httpStatus == http.StatusNotModified {
		return nil
	}

	if firstRun {
		logger.Debug("First fetch for feed '%s', logging %d articles without sending notifications",
			feed.ID, len(parsedFeed.Items))
		s.logItems(feed, parsedFeed.Items)
		return nil
	}

	return s.processArticles(feed, parsedFeed.Items)
}

// fetchFeed retrieves and parses a feed from its URL.
func (s *Service) fetchFeed(feed *config.Feed, metadata *db.Feed) (*gofeed.Feed, int, error) {
	req, err := http.NewRequestWithContext(s.ctx, "GET", feed.URL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	if metadata.ETag != "" {
		req.Header.Add("If-None-Match", metadata.ETag)
	} else if metadata.LastModified != "" {
		req.Header.Add("If-Modified-Since", metadata.LastModified)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if etag := resp.Header.Get("ETag"); etag != "" {
		metadata.ETag = etag
	}

	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		metadata.LastModified = lastModified
	}

	cacheControl := resp.Header.Get("Cache-Control")
	metadata.MaxAge = parseMaxAge(cacheControl, 14400) // 4 hours max

	switch resp.StatusCode {
	case http.StatusOK:
		parsedFeed, err := s.parser.Parse(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed to parse feed: %w", err)
		}
		return parsedFeed, resp.StatusCode, nil
	case http.StatusNotModified:
		return nil, http.StatusNotModified, nil
	default:
		return nil, resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// processArticles handles new articles in a feed and sends notifications.
func (s *Service) processArticles(feed *config.Feed, articles []*gofeed.Item) error {
	notifierInstance := s.getNotifierForFeed(feed)

	for _, item := range articles {
		articleID := s.getArticleID(item)
		if articleID == "" {
			continue
		}
		if s.db.IsArticleNew(feed.ID, articleID) {
			if err := notifierInstance.Notify(feed, item); err != nil {
				log.Printf("Failed to send notification for '%s': %v", articleID, err)
				continue
			}
			s.db.LogArticle(feed.ID, articleID)
		}
	}

	return nil
}
