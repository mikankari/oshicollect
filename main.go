package main

import (
	"fmt"
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
			`#みずえなで50音`,
		}, " OR ") + `) filter:links -filter:replies -filter:retweets`,
		ResultType: "recent",
		Count: 100,
		SinceID: lastSinceID,
		TweetMode: "extended",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%d / %d\n", len(search.Statuses), search.Metadata.Count)

	for i := len(search.Statuses) - 1; i >= 0; i-- {
		item := search.Statuses[i]

		fmt.Println(strings.ReplaceAll(item.FullText, "\n", " "))
		fmt.Printf("%d likes from %s\n", item.FavoriteCount, "https://twitter.com/" + item.User.ScreenName + "/status/" + item.IDStr)

		if ! func (item twitter.Tweet) (bool) {
			// テキストのみの引用 RT は引用元で見る。引用元を最近拾ったときは拾わない
			if item.QuotedStatus != nil && len(item.Entities.Urls) == 1 && len(item.Entities.Media) == 0 {
				item := item.QuotedStatus

				var retweeted bool = false
				for _, laterItem := range laterTweets {
					if laterItem.RetweetedStatus != nil && laterItem.RetweetedStatus.ID == item.ID {
						retweeted = true
						break
					}
				}
				if retweeted || len(item.Entities.Urls) == 0 && len(item.Entities.Media) == 0 {
					return false
				}
			}

			// 次のいずれかあれば、ふぁぼ数がなくても拾う
			hasContents := func (item twitter.Tweet) (bool) {
				for _, hashtag := range item.Entities.Hashtags {
					// タグ「みずえな」前方一致
					if strings.HasPrefix(hashtag.Text, "みずえな") {
						return true
					}
					// タグ「prsk_」前方一致
					if strings.HasPrefix(strings.ToLower(hashtag.Text), "prsk_") {
						return true
					}
					// タグ「1日1ニーゴ」
					if hashtag.Text == "1日1ニーゴ" {
						return true
					}
				}
				for _, url := range item.Entities.Urls {
					// リンク pixiv.net
					if strings.Contains(url.ExpandedURL, "//www.pixiv.net/") {
						return true
					}
				}
				// 説明が 15 文字程度以内
				if strings.Count(item.FullText, "") <= 40 {
					return true
				}
				return false
			}(item)

			// 本文に一致する
			// 「みずえな」とその他言語表記は 2 以上ふぁぼをもらえているか
			if strings.Contains(item.FullText, "みずえな") || strings.Contains(strings.ToLower(item.FullText), "mizuena") || strings.Contains(item.FullText, "미즈에나") {
				return hasContents || item.FavoriteCount >= 2
			}
			// 「瑞希」と「絵名」はグッズ情報などを避けるため、20 以上ふぁぼをもらえているか
			if strings.Contains(item.FullText, "瑞希") && strings.Contains(item.FullText, "絵名") {
				return hasContents || item.FavoriteCount >= 20
			}
			// 「mzen」は 2 以上ふぁぼもらえているか、加えてあいまいさ回避のため公式をフォローするか見る
			if strings.Contains(strings.ToLower(item.FullText), "mzen") {
				if hasContents || item.FavoriteCount >= 10 {
					return true
				} else if item.FavoriteCount >= 2 {
					result, _, err := client.Friendships.Show(&twitter.FriendshipShowParams{
						SourceID: item.User.ID,
						TargetID: 1158668053183266816,
					})
					if err != nil {
						fmt.Println(err)
					} else if result.Source.Following {
						return true
					}
				}
			}

			return false
		}(item) {
			fmt.Println("  -> Dont retweet")
			continue
		}

		if os.Getenv("APP_ENV") != "production" {
			fmt.Println("  -> Retweeting")
			continue
		}

		_, _, err = client.Statuses.Retweet(search.Statuses[i].ID, nil)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("  -> Retweeted")
		}

		time.Sleep(1 * time.Second)
	}
}
