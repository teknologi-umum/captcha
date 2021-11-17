package analytics

import (
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type UserMap struct {
	UserID      int64     `db:"user_id"`
	Username    string    `db:"username"`
	DisplayName string    `db:"display_name"`
	Counter     int       `db:"counter"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	JoinedAt    time.Time `db:"joined_at"`
}

func ParseToUser(user *tb.User) UserMap {
	return UserMap{
		UserID:      int64(user.ID),
		DisplayName: user.FirstName + utils.ShouldAddSpace(user) + user.LastName,
		Username:    user.Username,
	}
}
 