package requests

import (
	"context"
	"fmt"

	"github.com/mmcdole/gofeed"
)

var fp = gofeed.NewParser()

func FetchFeed(ctx context.Context, feedUrl string) (*gofeed.Feed, error) {
	feed, err := fp.ParseURL(feedUrl)

	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}

	return feed, nil
}
