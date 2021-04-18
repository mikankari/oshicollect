package main

import (
	"log"
	"os"
	"strings"
	"time"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func main() {
	client := twitter.NewClient(
		oauth1.NewConfig(os.Getenv("TWITTER_CONSUMER_KEY"), os.Getenv("TWITTER_CONSUMER_SECRET")).Client(
			oauth1.NoContext, oauth1.NewToken(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_SECRET"))))

	// 前回実行時から今までのツイートをざっくり得る
	var lastSinceID int64 = 0
	laterTweets, _, err := client.Timelines.UserTimeline(nil)
	for _, item := range laterTweets {
		if item.RetweetedStatus != nil {
			lastSinceID = item.RetweetedStatus.ID
			break
		}
	}
	if err != nil {
		log.Fatalln(err);
	}
	search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
		Query: `("みずえな" OR mzen) filter:links -filter:replies`,
		ResultType: "recent",
		Count: 100,
		SinceID: lastSinceID,
	})
	if err != nil {
		log.Fatalln(err);
	}
	if len(search.Statuses) == 100 {
		log.Println("may overflow")
	}

	for i := len(search.Statuses) - 1; i >= 0; i-- {
		item := search.Statuses[i]

		// RT は除く
		if item.RetweetedStatus != nil {
			continue
		}
		// 本文で一致する。mzen はあいまいさ回避のため仮にふぁぼ数も見る
		if ! (strings.Contains(item.Text, "みずえな") || (strings.Contains(item.Text, "mzen") && item.FavoriteCount >= 10)) {
			continue
		}

		log.Println(item.Text)
		log.Println("https://twitter.com/" + item.User.ScreenName + "/status/" + item.IDStr)

		_, _, err = client.Statuses.Retweet(item.ID, nil)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Retweeted")
		}

		time.Sleep(1 * time.Second)
	}
}
