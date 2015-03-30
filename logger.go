package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"log"
	"time"
)

func main() {
	log.Print("Running updated version")
	anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_TOKEN_SECRET)

	d := 60 * time.Second
	api.EnableThrottling(d, 4)
	api.SetLogger(anaconda.BasicLogger)

	log.Printf("Rate limiting with a token added every %s", d.String())

	followers_pages := api.GetFollowersListAll(nil)

	i := 0
	for page := range followers_pages {
		if page.Error != nil {
			log.Printf("ERROR: received error from GetFollowersListAll: %s", page.Error)
		}

		followers := page.Followers
		for _, follower := range followers {
			fmt.Printf("%+v\n", follower.ScreenName)
		}
		i++
	}
	log.Printf("Finished logging all %d followers -- exiting", i)
}
