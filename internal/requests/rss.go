package requests

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
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

func FetchFeed(ctx context.Context, feedUrl string) (RSSFeed, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)

	if err != nil {
		return RSSFeed{}, fmt.Errorf("err creating request: %w", err)
	}

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		return RSSFeed{}, fmt.Errorf("err requesting: %w", err)
	}

	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return RSSFeed{}, fmt.Errorf("err reading response bytes: %w", err)
	}

	var feed RSSFeed

	err = xml.Unmarshal(bytes, &feed)
	if err != nil {
		return RSSFeed{}, fmt.Errorf("err unmashaling xml: %w", err)
	}

	// Clean the fetched data
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for _, item := range feed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return feed, nil
}
