package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
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
	rssFeed.Channel.Title = html.EscapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.EscapeString(rssFeed.Channel.Description)
	for _, item := range rssFeed.Channel.Item {
		// item.Title = html.EscapeString(item.Title)
		fmt.Printf("%s\n", item.Title)
	}

}
