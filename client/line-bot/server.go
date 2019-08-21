// Copyright 2016 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/shinyamizuno1008/hashbill/server/db"
)

var SessionStore sessions.Store

const (
	inputFormat = "%s を入力してください。"
	ownerID     = "Udeadbeefdeadbeefdeadbeefdeadbeef"
	serverUrl   = "http://localhost:8000"
)

func main() {
	bot, err := linebot.New(
		Keys.ChannelSecret,
		Keys.ChannelToken,
	)
	if err != nil {
		log.Fatal(err)
	}

	cookieStore := sessions.NewCookieStore([]byte("something-very-secret"))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	SessionStore = cookieStore

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Source.UserID == ownerID {
				return
			}
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if message.Text == "会員登録" {
						if err := signupwithLINE(bot, event, w, req, message.Text); err != nil {
							log.Fatal(err)
						}
					}
					if message.Text == "whoami" {
						if err := showUser(bot, event); err != nil {
							log.Fatal(err)
						}
					}
					if message.Text == "イベント登録" {
						userSession, err := SessionStore.Get(req, event.Source.UserID)
						if err != nil {
							log.Fatal(err)
						}

						userSession.Options.MaxAge = -1
						err = userSession.Save(req, w)
						if err != nil {
							log.Fatal(err)
						}

						if err := registerEvent(bot, event, w, req, message.Text); err != nil {
							log.Fatal(err)
						}
					}
					userSession, err := SessionStore.Get(req, event.Source.UserID)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(userSession.Values["state"])
					if userSession.Values["state"] == "registeringEvent" {
						if err := registerEvent(bot, event, w, req, message.Text); err != nil {
							log.Fatal(err)
						}
					}
				}
			}
		}
	})

	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func signupwithLINE(bot *linebot.Client, event *linebot.Event, w http.ResponseWriter, req *http.Request, message string) *appError {
	userID := event.Source.UserID
	resFromLine, err := bot.GetProfile(userID).Do()
	if err != nil {
		return appErrorf(err, "could not get user profile: %v", err)
	}

	userName := resFromLine.DisplayName

	formData := url.Values{}
	formData.Set("userID", userID)
	formData.Add("userName", userName)

	_, err = http.PostForm(serverUrl+"/signup", formData)
	if err != nil {
		return appErrorf(err, "could not post user infor to the server: %v", err)
	}

	return nil
}

func showUser(bot *linebot.Client, event *linebot.Event) *appError {
	res, err := http.Get(serverUrl + "/user/" + event.Source.UserID)
	if err != nil {
		return appErrorf(err, "could not get user info from the server: %v", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return appErrorf(err, "could not read user reasponse body: %v", err)
	}

	var userInfo db.User
	json.Unmarshal(body, &userInfo)

	_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("あなたのIDは %s で、名前は %s ですね？", userInfo.UserID, userInfo.UserName))).Do()
	if err != nil {
		return appErrorf(err, "could not reply to user: %v", err)
	}

	return nil
}

func showEvents(bot *linebot.Client, event *linebot.Event) *appError {
	res, err := http.Get(serverUrl + "/events")
	if err != nil {
		return appErrorf(err, "could not get events info from the server: %v", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return appErrorf(err, "could not read events response body: %v", err)
	}

	var eventsInfo db.User
	json.Unmarshal(body, &eventsInfo)

	_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("イベント一覧")).Do()
	if err != nil {
		return appErrorf(err, "could not reply to user: %v", err)
	}

	return nil

}

func registerEvent(bot *linebot.Client, event *linebot.Event, w http.ResponseWriter, req *http.Request, message string) *appError {
	userSession, err := SessionStore.Get(req, event.Source.UserID)
	if err != nil {
		if err != nil {
			return appErrorf(err, "fail to create new session", err)
		}
	}

	// check if user register event in the first time.
	if userSession.Values["step"] == (interface{})(nil) {
		userSession, err = SessionStore.New(req, event.Source.UserID)
		userSession.Values["step"] = "begin"
	}

	switch userSession.Values["step"] {
	// register event steps; contain "eventName", "date", "deadline", "location", "members max", "lottery", "descriptin".
	case "begin":
		userSession.Values["state"] = "registeringEvent"
		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("イベントの登録を始めます。\n"+fmt.Sprintf(inputFormat, "イベント名"))).Do()
		if err != nil {
			return appErrorf(err, "could not reply to the user: %v", err)
		}

		userSession.Values["hostID"] = event.Source.UserID
		userSession.Values["step"] = "eventName"

		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not save session in begin step: %v", err)
		}
		return nil
	case "eventName":
		userSession.Values["eventName"] = message

		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf(inputFormat, "開催日時"))).Do()
		if err != nil {
			return appErrorf(err, "could not reply to user in eventName step: %v", err)
		}

		userSession.Values["step"] = "date"
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not to save session in eventName step: %v", err)
		}
		return nil
	case "date":
		userSession.Values["date"] = message

		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf(inputFormat, "締め切り"))).Do()
		if err != nil {
			return appErrorf(err, "could not reply to user in date step: %v", err)
		}

		userSession.Values["step"] = "deadline"
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not to save session in date step: %v", err)
		}
		return nil
	case "deadline":
		userSession.Values["deadline"] = message

		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf(inputFormat, "開催場所"))).Do()
		if err != nil {
			return appErrorf(err, "could not reply to user in deadline step: %v", err)
		}

		userSession.Values["step"] = "location"
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not to save session in deadline step: %v", err)
		}
		return nil
	case "location":
		userSession.Values["location"] = message

		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf(inputFormat, "参加者の上限"))).Do()
		if err != nil {
			return appErrorf(err, "could not reply to user in location step: %v", err)
		}

		userSession.Values["step"] = "membersMax"
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not to save session in location step: %v", err)
		}
		return nil
	case "membersMax":
		userSession.Values["membersMax"] = message

		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf(inputFormat, "抽選の有無"))).Do()
		if err != nil {
			return appErrorf(err, "could not reply to user in membersMax step: %v", err)
		}

		userSession.Values["step"] = "lottery"
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not to save session in membersMax step: %v", err)
		}
		return nil
	case "lottery":
		userSession.Values["lottery"] = message

		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf(inputFormat, "イベントの詳細"))).Do()
		if err != nil {
			return appErrorf(err, "could not reply to user in lottery step: %v", err)
		}

		userSession.Values["step"] = "description"
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not to save session in lottery step: %v", err)
		}
		return nil
	case "description":
		userSession.Values["description"] = message

		_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(`
		イベントの入力が完了しました。
		以下の内容で間違いありませんか？
		イベント名: `+userSession.Values["eventName"].(string)+`
		締め切り: `+userSession.Values["deadline"].(string)+`
		場所: `+userSession.Values["location"].(string)+`
		上限: `+userSession.Values["membersMax"].(string)+`
		抽選: `+userSession.Values["lottery"].(string)+`
		場所: `+userSession.Values["location"].(string))).Do()
		userSession.Values["step"] = "done"
		if err != nil {
			return appErrorf(err, "could not reply to user in description step: %v", err)
		}

		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not to save session in description step: %v", err)
		}
		return nil
	case "done":
		if message != "ok" {
			userSession.Values["step"] = "begin"
			err = userSession.Save(req, w)
			if err != nil {
				return appErrorf(err, "could not to save session in done step: %v", err)
			}

			_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("イベントの入力をはじめから入力し直してください。")).Do()
			if err != nil {
				return appErrorf(err, "could not reply to user in done step: %v", err)
			}
			return nil
		}
		membersMax, err := strconv.ParseInt(userSession.Values["membersMax"].(string), 10, 64)
		if err != nil {
			return appErrorf(err, "could not convert membersMax to int64 value: %v", err)
		}

		lottery, err := strconv.ParseBool(userSession.Values["lottery"].(string))
		if err != nil {
			return appErrorf(err, "could not convert lottery to boolean value: %v", err)
		}

		eventDetail := &db.Event{
			EventName:   userSession.Values["eventName"].(string),
			Date:        userSession.Values["date"].(string),
			Deadline:    userSession.Values["deadline"].(string),
			Location:    userSession.Values["location"].(string),
			MembersMax:  membersMax,
			Lottery:     lottery,
			Description: userSession.Values["description"].(string),
		}

		flexMessage, err := replyEventTicket(bot, eventDetail)
		if err != nil {
			return appErrorf(err, "%v", err)
		}

		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("イベントの登録が完了しました。\nチケットを表示します。"), flexMessage).Do()
		if err != nil {
			return appErrorf(err, "could not reply message to the user: %v", err)
		}

		userSession.Options.MaxAge = -1
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not save session in done step: %v", err)
		}
		return nil
	default:
		userSession.Values["step"] = "begin"
		err = userSession.Save(req, w)
		if err != nil {
			return appErrorf(err, "could not save session in %v", err)
		}
		return nil
	}
}

// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}

func replyEventTicket(bot *linebot.Client, event *db.Event) (*linebot.FlexMessage, error) {
	const eventFormat = `{
		"type": "bubble",
		"hero": {
		  "type": "image",
		  "url": "https://lh3.googleusercontent.com/BkvRJsjYiEjb0-XKuop2AurqFKLhhu_iIP06TrCTGAq180P9Briv8Avz8ncLp7bOmCs",
		  "size": "full",
		  "aspectRatio": "20:13",
		  "aspectMode": "cover",
		  "action": {
			"type": "uri",
			"uri": "http://linecorp.com/"
		  }
		},
		"body": {
		  "type": "box",
		  "layout": "vertical",
		  "spacing": "md",
		  "contents": [
			{
			  "type": "text",
			  "text": %s, 
			  "wrap": true,
			  "weight": "bold",
			  "gravity": "center",
			  "size": "xl"
			},
			{
			  "type": "box",
			  "layout": "vertical",
			  "margin": "lg",
			  "spacing": "sm",
			  "contents": [
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": "開催日時",
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": %s,
					  "wrap": true,
					  "size": "sm",
					  "color": "#666666",
					  "flex": 3
					}
				  ]
				},
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": "締め切り",
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": %s,
					  "wrap": true,
					  "size": "sm",
					  "color": "#666666",
					  "flex": 3
					}
				  ]
				},
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": 開催場所,
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": %s,
					  "wrap": true,
					  "color": "#666666",
					  "size": "sm",
					  "flex": 3
					}
				  ]
				},
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": 上限,
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": %s,
					  "wrap": tdue,
					  "color": "#666666",
					  "size": "sm",
					  "flex": 3
					}
				  ]
				},
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": "抽選",
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": %s,
					  "wrap": true,
					  "color": "#666666",
					  "size": "sm",
					  "flex": 3
					}
				  ]
				},
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": "説明",
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": %s,
					  "wrap": true,
					  "color": "#666666",
					  "size": "sm",
					  "flex": 3
					}
				  ]
				}
			  ]
			},
			{
			  "type": "box",
			  "layout": "vertical",
			  "margin": "xxl",
			  "contents": [
				{
				  "type": "spacer"
				},
				{
				  "type": "image",
				  "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/linecorp_code_withborder.png",
				  "aspectMode": "cover",
				  "size": "xl"
				},
				{
				  "type": "text",
				  "text": "この表示されたものがあなたが参加しようとしているイベントのチケットとなります。",
				  "color": "#aaaaaa",
				  "wrap": true,
				  "margin": "xxl",
				  "size": "xs"
				}
			  ]
			}
		  ]
		}
	  }`

	// parse members max (int64) and lottery (bool) to string.
	membersMax := strconv.FormatInt(event.MembersMax, 10)
	lottery := strconv.FormatBool(event.Lottery)

	eventJSON := []byte(fmt.Sprintf(eventFormat, event.EventName, event.Date, event.Deadline, event.Location, membersMax, lottery, event.Description))
	container, err := linebot.UnmarshalFlexMessageJSON(eventJSON)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal flex message: %v", err)
	}

	message := linebot.NewFlexMessage("Event ticket", container)
	return message, nil
}
