package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// GetFeed retrieves the metadata for a feed.
func (db *DB) GetFeed(feedID string) *Feed {
	var metadata Feed
	row := db.QueryRow("SELECT feed_id, etag, last_modified, max_age, last_checked FROM feeds WHERE feed_id = ?", feedID)
	err := row.Scan(&metadata.FeedID, &metadata.ETag, &metadata.LastModified, &metadata.MaxAge, &metadata.LastChecked)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		log.Fatalf("failed to write to database: %v", err)
	}
	return &metadata
}

// UpdateFeed updates the metadata for a feed.
func (db *DB) UpdateFeed(metadata *Feed) {
	_, err := db.Exec(`
        INSERT INTO feeds (feed_id, etag, last_modified, max_age, last_checked) 
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(feed_id) DO UPDATE SET
            etag = excluded.etag,
            last_modified = excluded.last_modified,
            max_age = excluded.max_age,
            last_checked = excluded.last_checked
    `, metadata.FeedID, metadata.ETag, metadata.LastModified, metadata.MaxAge, metadata.LastChecked)

	if err != nil {
		log.Fatalf("failed to write to database: %v", err)
	}
}

// IsArticleNew checks whether an article has been seen before.
func (db *DB) IsArticleNew(feedID string, articleID string) bool {
	var result int
	err := db.QueryRow(
		"SELECT 1 FROM articles WHERE feed_id = ? AND article_id = ? LIMIT 1",
		feedID, articleID,
	).Scan(&result)

	if err == sql.ErrNoRows {
		return true
	}

	if err != nil {
		log.Fatalf("failed to read from database: %v", err)
	}

	return false
}

// LogArticle logs an article after we've processed it and sent a notification.
func (db *DB) LogArticle(feedID string, articleID string) {
	_, err := db.Exec(
		"INSERT OR IGNORE INTO articles (feed_id, article_id) VALUES (?, ?)",
		feedID, articleID,
	)
	if err != nil {
		log.Fatalf("failed to write to database: %v", err)
	}
}
