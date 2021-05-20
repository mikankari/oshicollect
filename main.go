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
		Query: `(` + strings.Join([]string{
			`"みずえ"`,
			`"みず え"`,
			`"ずえな"`,
			`"ず えな"`,
			`#みずえな`,
			`mzen`,
			`mizuena`,
			`"미즈 에나"`,
			`(瑞希 絵名 -交換)`,
			`#みずえな25時ワンドロワンライ`,
		}, " OR ") + `) filter:links -filter:replies -filter:retweets`,
		ResultType: "recent",
		Count: 100,
		SinceID: lastSinceID,
		TweetMode: "extended",
	})
	if err != nil {
		log.Fatalln(err);
	}
	log.Printf("%d / %d", len(search.Statuses), search.Metadata.Count)

	for i := len(search.Statuses) - 1; i >= 0; i-- {
		item := search.Statuses[i]

		log.Println(strings.ReplaceAll(item.FullText, "\n", " "))
		log.Println("https://twitter.com/" + item.User.ScreenName + "/status/" + item.IDStr)

		if ! func (item twitter.Tweet) (bool) {
			// 引用 RT を除く
			if item.QuotedStatus != nil {
				return false
			}

			// ふぁぼがないのを除く
			if item.FavoriteCount < 1 {
				return false
			}

			// 本文に一致する
			if strings.Contains(item.FullText, "みずえな") || strings.Contains(strings.ToLower(item.FullText), "mizuena") || strings.Contains(item.FullText, "미즈에나") {
				return true
			}
			if strings.Contains(item.FullText, "瑞希") && strings.Contains(item.FullText, "絵名") {
				return true
			}
			// mzen はあいまいさ回避のため、加えてふぁぼ数または公式をフォローするか見る
			if strings.Contains(strings.ToLower(item.FullText), "mzen") {
				if item.FavoriteCount >= 10 {
					return true
				}
				result, _, err := client.Friendships.Show(&twitter.FriendshipShowParams{
					SourceID: item.User.ID,
					TargetID: 1158668053183266816,
				})
				if err != nil {
					log.Println(err)
				} else if result.Source.Following {
					return true
				}
			}

			return false
		}(item) {
			continue
		}

		log.Println("Retweeting")

		_, _, err = client.Statuses.Retweet(item.ID, nil)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Retweeted")
		}

		time.Sleep(1 * time.Second)
	}
}
