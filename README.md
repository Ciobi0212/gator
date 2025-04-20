# Gator - RSS Feed Aggregator

![Gator Logo](https://img.shields.io/badge/Gator-RSS%20Feed%20Aggregator-green)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)

Gator is a powerful command-line RSS feed aggregator built in Go that helps you follow and discover content from your favorite websites without leaving your terminal.

## Features

- 📚 **User Management**: Create and manage multiple user accounts
- 📡 **Feed Management**: Add, follow, and unfollow RSS feeds from any website
- 🔍 **Content Discovery**: Browse the latest posts from feeds you follow
- ⏱️ **Automatic Updates**: Aggregate content at your preferred intervals
- 🚀 **Fast & Lightweight**: Runs efficiently in your terminal
- ⚡ **Concurrent Processing**: Fetch multiple feeds simultaneously for better performance

## Installation

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/Ciobi0212/gator.git
   cd gator
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up the database (make sure PostgreSQL is running):
   ```bash
   # Install goose for migrations
   go install github.com/pressly/goose/v3/cmd/goose@latest
   
   # Run migrations
   goose -dir sql/schema postgres "postgres://username:password@localhost:5432/gator_db?sslmode=disable" up
   ```

4. Build the application:
   ```bash
   go build -o gator
   ```

## Quick Start

1. **Register a new user:**
   ```bash
   ./gator register john
   ```

2. **Add an RSS feed:**
   ```bash
   ./gator addfeed "Hacker News" https://news.ycombinator.com/rss
   ```

3. **Start the aggregator in the background:**
   ```bash
   ./gator agg 30m 5 &
   ```
   This will fetch new content every 30 minutes with 5 concurrent workers.

4. **Browse the latest posts:**
   ```bash
   ./gator browse 10
   ```
   This will display the 10 most recent posts.

## Usage Guide

### User Management

| Command | Description | Example |
|---------|-------------|---------|
| `register <username>` | Create a new user account | `./gator register john` |
| `login <username>` | Log in as an existing user | `./gator login john` |
| `users` | List all registered users | `./gator users` |

### Feed Management

| Command | Description | Example |
|---------|-------------|---------|
| `addfeed <name> <url>` | Add a new RSS feed and follow it | `./gator addfeed "Tech News" https://example.com/rss` |
| `feeds` | List all available feeds | `./gator feeds` |
| `follow <url>` | Follow an existing feed | `./gator follow https://example.com/rss` |
| `following` | List all feeds you're following | `./gator following` |
| `unfollow <url>` | Unfollow a feed | `./gator unfollow https://example.com/rss` |

### Content

| Command | Description | Example |
|---------|-------------|---------|
| `browse <limit>` | View posts from feeds you follow | `./gator browse 20` |

### System

| Command | Description | Example |
|---------|-------------|---------|
| `agg <interval> [concurrency]` | Start feed aggregation process | `./gator agg 1h 3` |
| | interval: time between fetches | |
| | concurrency: number of feeds to fetch in parallel (default: 1) | |
| `reset` | Delete all users and feeds (use with caution) | `./gator reset` |
| `help` | Display help information | `./gator help` |

## Example Workflow

Here's a complete workflow example:

```bash
# Register a user
./gator register sarah

# Add some tech news feeds
./gator addfeed "Hacker News" https://news.ycombinator.com/rss
./gator addfeed "The Verge" https://www.theverge.com/rss/index.xml
./gator addfeed "Wired" https://www.wired.com/feed/rss

# Start aggregating content in the background (every 15 minutes with 3 concurrent workers)
./gator agg 15m 3 &

# Check what feeds you're following
./gator following

# Browse the 15 most recent posts
./gator browse 15

# Stop following a feed
./gator unfollow https://www.theverge.com/rss/index.xml
```

## Tips & Tricks

- Run `./gator agg` in a separate terminal window or as a background process to continuously fetch new content
- For faster updates with many feeds, increase the concurrency parameter (e.g., `./gator agg 10m 10`)
- Use `./gator help` to see all available commands
- Set up a cronjob to run the aggregator automatically at system startup

## Troubleshooting

If you encounter issues with certain feeds not parsing correctly, try these solutions:

1. Verify the feed URL is correct and accessible
2. Check that the feed is a valid RSS format
