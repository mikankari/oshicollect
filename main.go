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
		Query: `("みずえ" OR "みず え" OR "ずえな" OR "ず えな" OR #みずえな OR mzen OR mizuena OR "미즈 에나" OR #みずえな25時ワンドロワンライ) filter:links -filter:replies -filter:retweets`,
		ResultType: "recent",
		Count: 100,
		SinceID: lastSinceID,
		TweetMode: "extended",
	})
	if err != nil {
		log.Fatalln(err);
	}
	if len(search.Statuses) == 100 {
		log.Println("may overflow")
	}

	for i := len(search.Statuses) - 1; i >= 0; i-- {
		item := search.Statuses[i]

		// 引用 RT を除く
		if item.QuotedStatus != nil {
			continue
		}

		// 本文に一致する
		if ! (strings.Contains(item.FullText, "みずえな") || strings.Contains(strings.ToLower(item.FullText), "mizuena") || strings.Contains(item.FullText, "미즈에나")) {
			if ! strings.Contains(strings.ToLower(item.FullText), "mzen") {
				continue
			}
			// mzen はあいまいさ回避のため、ふぁぼ数または公式をフォローするか見る
			if item.FavoriteCount < 10 {
				result, _, err := client.Friendships.Show(&twitter.FriendshipShowParams{
					SourceID: item.User.ID,
					TargetID: 1158668053183266816,
				})
				if err != nil {
					log.Println(err)
					continue
				}
				if ! result.Source.Following {
					continue
				}
			}
		}

		log.Println(item.FullText)
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
