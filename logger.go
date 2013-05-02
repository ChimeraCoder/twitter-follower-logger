package main

import (
    "github.com/ChimeraCoder/anaconda"
    "log"
    "net/url"
    "time"
)


func main(){
    anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
    anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
    api := anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN,TWITTER_ACCESS_TOKEN_SECRET)

    v := url.Values{}

    users := []anaconda.TwitterUser{}

    next_cursor := "-1"
    for {
        v.Set("cursor", next_cursor)
        c, err := api.GetFollowersList(v)
        if err != nil{
            panic(err)
        }

        users = append(users, c.Users...)

        next_cursor = c.Next_cursor_str
        if next_cursor == "0" {
            break
        } else {
            log.Printf("Calling again with cursor %v", next_cursor)
        }
        time.Sleep(5 * time.Second)
    }



    for _, user := range users {
        log.Printf("%+v", *user.Screen_name)
    }
}
