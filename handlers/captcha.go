package handlers

import (
	"context"
	"encoding/json"
	"strconv"
	"teknologi-umum-bot/utils"
	"time"

	"github.com/allegro/bigcache/v3"
	"go.mongodb.org/mongo-driver/bson"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Captcha struct {
	Question   string `json:"question" bson:"question"`
	Answer     string `json:"answer" bson:"answer"`
	ContentURL string `json:"contenturl" bson:"contenturl"`
}

func (d *Dependencies) FetchCaptchas() ([]Captcha, error) {
	ctx, cancel := context.WithTimeout(d.Context, 30*time.Second)
	defer cancel()

	captchaCollection := d.Mongo.Collection("captcha")
	var captcha []Captcha
	err := captchaCollection.Find(ctx, bson.M{}).All(&captcha)
	if err != nil {
		return []Captcha{}, err
	}

	return captcha, nil
}

func (d *Dependencies) GetRandomCaptcha() (Captcha, error) {
	var captcha []Captcha
	captchas, err := d.Cache.Get("captchas")
	if err != nil && err != bigcache.ErrEntryNotFound {
		return Captcha{}, err
	}

	if err == bigcache.ErrEntryNotFound {
		tempCaptcha, err := d.FetchCaptchas()
		if err != nil {
			return Captcha{}, err
		}

		captcha = tempCaptcha
		captchaBytes, err := json.Marshal(tempCaptcha)
		if err != nil {
			return Captcha{}, err
		}

		err = d.Cache.Set("captchas", captchaBytes)
		if err != nil {
			return Captcha{}, err
		}
	} else {
		json.Unmarshal(captchas, &captcha)
	}

	randInt, err := utils.GenerateRandomNumber(len(captcha))
	if err != nil {
		return Captcha{}, err
	}

	return captcha[randInt], nil
}

func (d *Dependencies) CaptchaUserJoin(m *tb.Message) {
	// Pick a random photo
	captcha, err := d.GetRandomCaptcha()
	if err != nil {
		panic(err)
	}

	captchaImage := &tb.Photo{File: tb.FromURL(captcha.ContentURL)}
	msgImage, err := d.Bot.Send(m.Sender, captchaImage)
	if err != nil {
		panic(err)
	}

	msgQuestion, err := d.Bot.Send(m.Sender, captcha.Question)
	if err != nil {
		panic(err)
	}

	// Start expiry in the key form of "<sender telegram ID>:expiry"
	err = d.Cache.Set(strconv.Itoa(m.Sender.ID)+":expiry", []byte(strconv.FormatInt(time.Now().Unix()+int64(5*time.Minute), 10)))
	if err != nil {
		panic(err)
	}
	// Store theh answer in the key form of "<sender telegram ID>:answer"
	err = d.Cache.Set(strconv.Itoa(m.Sender.ID)+":answer", []byte(captcha.Answer))
	if err != nil {
		panic(err)
	}

	expiryTimer := time.AfterFunc(5*time.Minute, func() {
		err := d.Cache.Delete(strconv.Itoa(m.Sender.ID) + ":expiry")
		if err != nil {
			panic(err)
		}

		err = d.Cache.Delete(strconv.Itoa(m.Sender.ID) + ":answer")
		if err != nil {
			panic(err)
		}

		_, err = d.Bot.Send(m.Sender,
			"<a href=\"tg://user?id="+strconv.Itoa(m.Sender.ID)+"\">"+m.Sender.FirstName+" "+m.Sender.LastName+"</a> didn't solve the captcha. Aight, time to kick them.",
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
			})
		if err != nil {
			panic(err)
		}

		err = d.Bot.Ban(m.Chat, &tb.ChatMember{
			RestrictedUntil: time.Now().Unix() + int64(1*time.Hour),
			User:            m.Sender,
		}, true)
		if err != nil {
			panic(err)
		}

		err = d.Bot.Delete(msgImage)
		if err != nil {
			panic(err)
		}

		err = d.Bot.Delete(msgQuestion)
		if err != nil {
			panic(err)
		}
	})

	d.Timers <- CaptchaTimer{
		strconv.Itoa(m.Sender.ID): CaptchaTimerConfig{
			Timer:           expiryTimer,
			Sender:          m.Sender,
			Chat:            m.Chat,
			MessageImage:    msgImage,
			MessageQuestion: msgQuestion,
		},
	}
}

func (d *Dependencies) CaptchaUserMessage(m *tb.Message) {
	// Listen to user message, check if current message matches
	_, err := d.Cache.Get(strconv.Itoa(m.Sender.ID) + ":expiry")
	if err != nil {
		return
	}

}
