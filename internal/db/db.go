package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Feed represents the metadata for a feed.
type Feed struct {
	FeedID       string `db:"feed_id"`
	ETag         string `db:"etag"`
	LastModified string `db:"last_modified"`
	MaxAge       int64  `db:"max_age"`
	LastChecked  int64  `db:"last_checked"`
}

// Article represents an article in a feed.
type Article struct {
	FeedID    string `db:"feed_id"`
	ArticleID string `db:"article_id"`
}

// DB holds the database information.
type DB struct {
	*sql.DB
}

func Open(dbFile string) (*DB, error) {
	dir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS feeds (
		    feed_id TEXT NOT NULL PRIMARY KEY,
		    etag TEXT,
		    last_modified TEXT,
		    max_age INTEGER,
		    last_checked INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS articles (
		    feed_id TEXT NOT NULL,
		    article_id TEXT NOT NULL,
		    PRIMARY KEY(feed_id, article_id)
		);
	`); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	d := &DB{DB: db}
	return d, nil
}
