package twitter

import (
	"fmt"
	"log"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Config struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
	UserID         int64
	Filters        []string
}

type Twitter struct {
	client  *twitter.Client
	stream  *twitter.Stream
	ch      chan string
	UserID  int64
	Filters []string
}

func NewTwitter(config *Config, ch chan string) *Twitter {
	oconfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)

	return &Twitter{
		client:  twitter.NewClient(oconfig.Client(oauth1.NoContext, oauth1.NewToken(config.AccessToken, config.AccessSecret))),
		stream:  nil,
		ch:      ch,
		UserID:  config.UserID,
		Filters: config.Filters,
	}
}

func (t Twitter) Start() error {
	demux := twitter.NewSwitchDemux()
	demux.Tweet = t.handleTweet
	demux.Event = t.handleEvent

	log.Println("🐔 Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Follow:        []string{fmt.Sprintf("%d", t.UserID)},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := t.client.Streams.Filter(filterParams)
	if err == nil {
		t.stream = stream
		go demux.HandleChan(t.stream.Messages)
	}
	return err
}

func (t Twitter) Stop() {
	log.Println("🐔 Stopping Stream...")
	if t.stream != nil {
		t.stream.Stop()
		t.stream = nil
	}
}

func (t Twitter) handleTweet(tweet *twitter.Tweet) {
	log.Printf("🐔🏃 Processing tweet %v\n", tweet.Text)
	if t.filterTweet(tweet) {
		log.Println("🐔🏃 Sending to channel")
		t.ch <- fmt.Sprintf("https://twitter.com/%s/status/%d", tweet.User.ScreenName, tweet.ID)
	} else {
		log.Println("🐔🏃 Tweet is NOT worthy!")
	}
}

func (t Twitter) filterTweet(tweet *twitter.Tweet) bool {
	if tweet.User.ID == t.UserID {
		tweetText := strings.ToLower(tweet.Text)
		for _, f := range t.Filters {
			if strings.Contains(tweetText, f) {
				return true
			}
		}
	}
	return false
}

func (t Twitter) handleEvent(event *twitter.Event) {
	log.Printf("🐔 %#v\n", event)
}
