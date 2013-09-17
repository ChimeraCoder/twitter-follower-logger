package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"log"
	"net/url"
	"time"
)

func main() {
	anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_TOKEN_SECRET)

	v := url.Values{}

	users := []anaconda.TwitterUser{}

	next_cursor := "-1"
	for {
		v.Set("cursor", next_cursor)
		c, err := api.GetFollowersList(v)

		//TODO distinguish between rate limiting errors and other errors
		if err != nil {
			log.Printf("ERROR: %v", err)
			time.Sleep(90 * time.Second)
			continue //Retry
		}

		users = append(users, c.Users...)
		log.Printf("Appended %d users", len(c.Users))
		go func(us []anaconda.TwitterUser) {
			for _, user := range us {
				log.Printf("Appended %+v", *user.Screen_name)
			}
		}(c.Users)

		next_cursor = c.Next_cursor_str
		if next_cursor == "0" {
			break
		} else {
			log.Printf("Calling again with cursor %v", next_cursor)
		}
		time.Sleep(60 * time.Second)
	}

	for _, user := range users {
		fmt.Printf("%+v", *user.Screen_name)
	}
}
