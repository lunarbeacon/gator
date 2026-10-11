package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"gator/internal/database"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	// 1. Create a context with a timeout
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// 2. Create the request with the context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "gator")

	// 3. Execute the request using http.Client.Do
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed
	if err := xml.Unmarshal(data, &rssFeed); err != nil {
		return nil, err
	}

	return &rssFeed, nil
}

func scrapeFeeds(s *state) {
	ctx := context.Background()

	// get the next feed
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		fmt.Printf("Failed to fetch feed: %v", err)
	}
	fmt.Printf("getting feed %s\n", feed.Name)
	// mark it as fetched
	_, err = s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		fmt.Printf("Failed to mark feed: %v", err)
	}

	// get the feed by its url
	var rssFeed *RSSFeed
	rssFeed, err = fetchFeed(ctx, feed.Url)
	if err != nil {
		fmt.Printf("Failed to fetch feed: %v", err)
	}

	// load the feed items into the posts table in the database
	for _, item := range rssFeed.Channel.Item {
		parms := database.CreatePostParams{}
		parms.ID = uuid.New()
		parms.CreatedAt = time.Now()
		parms.UpdatedAt = time.Now()
		parms.Title = item.Title
		parms.Url = item.Link
		if len(item.Description) == 0 {
			parms.Description = sql.NullString{}
		} else {
			parms.Description = sql.NullString{String: item.Description, Valid: true}
		}
		// try to parse the pub date two ways
		t, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			t, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", item.PubDate)
			if err != nil {
				fmt.Printf("Failed to parse pub date: %s, %v", item.PubDate, err)
				parms.PublishedAt = sql.NullTime{}
			}
		} else {
			parms.PublishedAt = sql.NullTime{Time: t, Valid: true}
		}
		parms.FeedID = feed.ID
		_, _ = s.db.CreatePost(ctx, parms)

	}

}
