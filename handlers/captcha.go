package handlers

import (
	"encoding/json"
	"teknologi-umum-bot/utils"

	"github.com/allegro/bigcache/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Captcha struct {
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	ContentURL string `json:"contenturl"`
}

func (d *Dependencies) CaptchaUserJoin(m *tb.Message) {
	// Pick a random photo
	captchas, err := d.Cache.Get("captchas")
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			// Fetch captchas from database
			// Panic() is a placeholder so I don't have a warning on my IDE.
			panic(err)
		}
		// PANIC! No, we shouldn't be panic. Just throw error and let the user know.
		// "Lucky bastard, we got an error. You may enter without captcha."
	}

	var captcha []Captcha
	err = json.Unmarshal(captchas, &captcha)
	if err != nil {
		// throw error, how? don't know might think about it later
		panic(err)
	}

	// Get random captcha from the captcha
	randInt, err := utils.GenerateRandomNumber(len(captcha))
	if err != nil {
		// Remember, this is still a placeholder
		panic(err)
	}

	// This algorithm seems sucks lol
	// Might change it later
	selectedCaptcha := captcha[randInt]
	d.Bot.Send(m.Sender, selectedCaptcha)
	d.Bot.Send(m.Sender, "Ini nanti diganti foto")

}

func (d *Dependencies) CaptchaUserMessage(m *tb.Message) {
	// Listen to user message, check if current message matches
}
