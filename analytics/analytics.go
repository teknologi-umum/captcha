package analytics

import (
	"github.com/allegro/bigcache/v3"
	"github.com/jmoiron/sqlx"
	tb "gopkg.in/telebot.v3"
)

// Dependency is the dependency injection struct
// for the analytics package.
type Dependency struct {
	Memory   *bigcache.BigCache
	Bot      *tb.Bot
	DB       *sqlx.DB
	TeknumID string
}

// HourMapper is meant to use for mapping a time.Hour() to a string
var HourMapper = []string{
	"zero_hour", "one_hour", "two_hour", "three_hour", "four_hour", "five_hour",
	"six_hour", "seven_hour", "eight_hour", "nine_hour", "ten_hour", "eleven_hour",
	"twelve_hour", "thirteen_hour", "fourteen_hour", "fifteen_hour", "sixteen_hour",
	"seventeen_hour", "eighteen_hour", "nineteen_hour", "twenty_hour", "twentyone_hour",
	"twentytwo_hour", "twentythree_hour",
}

// HourlyMap contains the struct surrounding the hourly analytics.
type HourlyMap struct {
	TodaysDate      string `json:"todays_date" db:"todays_date"`
	ZeroHour        int    `json:"zero_hour" db:"zero_hour"`
	OneHour         int    `json:"one_hour" db:"one_hour"`
	TwoHour         int    `json:"two_hour" db:"two_hour"`
	ThreeHour       int    `json:"three_hour" db:"three_hour"`
	FourHour        int    `json:"four_hour" db:"four_hour"`
	FiveHour        int    `json:"five_hour" db:"five_hour"`
	SixHour         int    `json:"six_hour" db:"six_hour"`
	SevenHour       int    `json:"seven_hour" db:"seven_hour"`
	EightHour       int    `json:"eight_hour" db:"eight_hour"`
	NineHour        int    `json:"nine_hour" db:"nine_hour"`
	TenHour         int    `json:"ten_hour" db:"ten_hour"`
	ElevenHour      int    `json:"eleven_hour" db:"eleven_hour"`
	TwelveHour      int    `json:"twelve_hour" db:"twelve_hour"`
	ThirteenHour    int    `json:"thirteen_hour" db:"thirteen_hour"`
	FourteenHour    int    `json:"fourteen_hour" db:"fourteen_hour"`
	FifteenHour     int    `json:"fifteen_hour" db:"fifteen_hour"`
	SixteenHour     int    `json:"sixteen_hour" db:"sixteen_hour"`
	SeventeenHour   int    `json:"seventeen_hour" db:"seventeen_hour"`
	EighteenHour    int    `json:"eighteen_hour" db:"eighteen_hour"`
	NineteenHour    int    `json:"nineteen_hour" db:"nineteen_hour"`
	TwentyHour      int    `json:"twenty_hour" db:"twenty_hour"`
	TwentyOneHour   int    `json:"twentyone_hour" db:"twentyone_hour"`
	TwentyTwoHour   int    `json:"twentytwo_hour" db:"twentytwo_hour"`
	TwentyThreeHour int    `json:"twentythree_hour" db:"twentythree_hour"`
}
