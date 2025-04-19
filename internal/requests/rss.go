package requests

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
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

	var feed RSSFeed

	decoder := xml.NewDecoder(res.Body)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false

	err = decoder.Decode(&feed)
	if err != nil {
		return RSSFeed{}, fmt.Errorf("err decoding xml: %w", err)
	}

	// Clean the fetched data
	fmt.Println(feed)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for _, item := range feed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return feed, nil
}
