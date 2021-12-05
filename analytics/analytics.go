package analytics

// On this package we have 2 main keys on redis:
// analytics:hour and analytics:counter

import (
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Dependency struct {
	Memory *bigcache.BigCache
	Bot    *tb.Bot
	Logger *sentry.Client
	DB     *sqlx.DB
}

// For mapping a time.Hour() to a string
var HourMapper []string = []string{
	"zero_hour", "one_hour", "two_hour", "three_hour", "four_hour", "five_hour",
	"six_hour", "seven_hour", "eight_hour", "nine_hour", "ten_hour", "eleven_hour",
	"twelve_hour", "thirteen_hour", "fourteen_hour", "fifteen_hour", "sixteen_hour",
	"seventeen_hour", "eighteen_hour", "nineteen_hour", "twenty_hour", "twenty_one_hour",
	"twenty_two_hour", "twenty_three_hour",
}

type HourlyMap struct {
	TodaysDate      time.Time `json:"todays_date" db:"todays_date"`
	ZeroHour        int       `json:"zero_hour" db:"zero_hour"`
	OneHour         int       `json:"one_hour" db:"one_hour"`
	TwoHour         int       `json:"two_hour" db:"two_hour"`
	ThreeHour       int       `json:"three_hour" db:"three_hour"`
	FourHour        int       `json:"four_hour" db:"four_hour"`
	FiveHour        int       `json:"five_hour" db:"five_hour"`
	SixHour         int       `json:"six_hour" db:"six_hour"`
	SevenHour       int       `json:"seven_hour" db:"seven_hour"`
	EightHour       int       `json:"eight_hour" db:"eight_hour"`
	NineHour        int       `json:"nine_hour" db:"nine_hour"`
	TenHour         int       `json:"ten_hour" db:"ten_hour"`
	ElevenHour      int       `json:"eleven_hour" db:"eleven_hour"`
	TwelveHour      int       `json:"twelve_hour" db:"twelve_hour"`
	ThirteenHour    int       `json:"thirteen_hour" db:"thirteen_hour"`
	FourteenHour    int       `json:"fourteen_hour" db:"fourteen_hour"`
	FifteenHour     int       `json:"fifteen_hour" db:"fifteen_hour"`
	SixteenHour     int       `json:"sixteen_hour" db:"sixteen_hour"`
	SeventeenHour   int       `json:"seventeen_hour" db:"seventeen_hour"`
	EighteenHour    int       `json:"eighteen_hour" db:"eighteen_hour"`
	NineteenHour    int       `json:"nineteen_hour" db:"nineteen_hour"`
	TwentyHour      int       `json:"twenty_hour" db:"twenty_hour"`
	TwentyOneHour   int       `json:"twenty_one_hour" db:"twenty_one_hour"`
	TwentyTwoHour   int       `json:"twenty_two_hour" db:"twenty_two_hour"`
	TwentyThreeHour int       `json:"twenty_three_hour" db:"twenty_three_hour"`
}
