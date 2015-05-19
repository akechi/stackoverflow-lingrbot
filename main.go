package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-lingr"
)

type resp struct {
	Items []struct {
		Tags  []string `json:"tags"`
		Owner struct {
			Reputation   int    `json:"reputation"`
			UserID       int    `json:"user_id"`
			UserType     string `json:"user_type"`
			ProfileImage string `json:"profile_image"`
			DisplayName  string `json:"display_name"`
			Link         string `json:"link"`
		} `json:"owner"`
		IsAnswered       bool   `json:"is_answered"`
		ViewCount        int    `json:"view_count"`
		AnswerCount      int    `json:"answer_count"`
		Score            int    `json:"score"`
		LastActivityDate int    `json:"last_activity_date"`
		CreationDate     int    `json:"creation_date"`
		QuestionID       int    `json:"question_id"`
		Link             string `json:"link"`
		Title            string `json:"title"`
	} `json:"items"`
	HasMore        bool `json:"has_more"`
	QuotaMax       int  `json:"quota_max"`
	QuotaRemaining int  `json:"quota_remaining"`
}

var re = regexp.MustCompile(`^stackoverflow (.+)$`)

func defaultAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":80"
	}
	return ":" + port
}

var addr = flag.String("addr", defaultAddr(), "server address")

func main() {
	flag.Parse()

	r := gin.Default()

	r.POST("/", func(c *gin.Context) {
		var status lingr.Status
		if c.EnsureBody(&status) {
			urls := ""
			for _, event := range status.Events {
				message := event.Message
				if message == nil {
					continue
				}
				if !re.MatchString(message.Text) {
					continue
				}
				question := re.FindStringSubmatch(message.Text)[1]
				params := url.Values{}
				params.Add("intitle", question)
				params.Add("site", "stackoverflow")
				params.Add("sort", "activity")
				params.Add("order", "desc")
				res, err := http.Get("https://api.stackexchange.com/2.2/search?" + params.Encode())
				println("https://api.stackexchange.com/2.2/search?" + params.Encode())
				if err != nil {
					println(err.Error())
					continue
				}
				defer res.Body.Close()
				var resp resp
				if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
					println(err.Error())
					continue
				}
				println(len(resp.Items))
				for _, item := range resp.Items {
					println(item.Link)
					s := fmt.Sprintf("%s\n%s\n", item.Title, item.Link)
					if len(urls+s) > 1000 {
						break
					}
					urls += s
				}
			}
			println(urls)
			c.String(200, urls)
			return
		}
		c.String(200, "")
	})
	r.Run(*addr)
}
