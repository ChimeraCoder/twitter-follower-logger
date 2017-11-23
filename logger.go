package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ChimeraCoder/anaconda"

	"github.com/DataDog/datadog-go/statsd"

	"cloud.google.com/go/storage"
)

var VERSION = "unset"

var TWITTER_CONSUMER_KEY = os.Getenv("TWITTER_CONSUMER_SECRET")
var TWITTER_CONSUMER_SECRET = os.Getenv("TWITTER_CONSUMER_SECRET")
var TWITTER_ACCESS_TOKEN = os.Getenv("TWITTER_ACCESS_TOKEN")
var TWITTER_ACCESS_TOKEN_SECRET = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

var THROTTLE_SECONDS = os.Getenv("THROTTLE_SECONDS")

func main() {

	stats, err := statsd.NewBuffered("127.0.0.1:8126", 1024)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Running version %s", VERSION)
	anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_TOKEN_SECRET)

	d := 120 * time.Second

	if THROTTLE_SECONDS != "" {
		i, err := strconv.Atoi(THROTTLE_SECONDS)
		if err != nil {
			log.Printf("Could not parse %s: %s", THROTTLE_SECONDS, err)
		} else {
			d = time.Duration(i) * time.Second
		}
	}

	api.EnableThrottling(d, 4)
	api.SetLogger(anaconda.BasicLogger)

	stats.Count("twitterfollowerlogger.init", 1, []string{fmt.Sprintf("duration:%s", d.String())}, 1.0)
	log.Printf("Rate limiting with a token added every %s", d.String())

	followers_pages := api.GetFollowersListAll(nil)

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Could not create storage client: %s", err)
	}

	bkt := client.Bucket("twitter-follower-logger")

	start := time.Now().Unix()
	recordFilename := strconv.Itoa(int(start))
	logFilename := fmt.Sprintf("%s.log", recordFilename)

	logF := bkt.Object(logFilename).NewWriter(ctx)
	defer logF.Close()

	log.SetOutput(io.MultiWriter(logF, os.Stderr))

	log.Printf("logging to %s", logFilename)

	recordF := bkt.Object(recordFilename).NewWriter(ctx)
	defer recordF.Close()

	log.Printf("writing records to %s", recordFilename)

	i := 0
	count := 0
	for page := range followers_pages {
		stats.Count("twitterfollowerlogger.page", 1, nil, 1.0)
		if page.Error != nil {
			err = page.Error
			log.Printf("ERROR: received error from GetFollowersListAll: %s", page.Error)
			stats.Count("twitterfollowerlogger.page.errors", 1, []string{fmt.Sprintf("error:%s", page.Error)}, 1.0)
		} else {
			err = nil
		}

		followers := page.Followers
		for _, follower := range followers {
			_, err := fmt.Fprintf(recordF, "%+v\n", follower.ScreenName)
			if err != nil {
				stats.Count("twitterfollowerlogger.record.write_error", int64(count), []string{fmt.Sprintf("error:%s", err.Error())}, 1.0)
				log.Fatalf("Error writing to record file %s", err)

			}
		}
		i++
		count += len(followers)
		stats.Count("twitterfollowerlogger.page.page_length", int64(count), nil, 1.0)
	}
	if err != nil {
		stats.Count("twitterfollowerlogger.app.errors", 1, []string{fmt.Sprintf("error:%s", err.Error())}, 1.0)
		log.Fatalf("Exiting due to error %s", err)
	}
	stats.Gauge("twitterfollowerlogger.followers_total", float64(i), nil, 1.0)
	log.Printf("Finished logging all %d followers -- exiting", count)
}
