package analytics

import (
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// UserMap contains a data of a user.
type UserMap struct {
	GroupID     int64     `json:"group_id,omitempty" db:"group_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Username    string    `json:"username,omitempty" db:"username" redis:"username"`
	DisplayName string    `json:"display_name,omitempty" db:"display_name" redis:"display_name"`
	Counter     int       `json:"counter" db:"counter" redis:"counter"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

// ParseToUser converts the tb.Message struct into a UserMap struct.
func ParseToUser(m *tb.Message) UserMap {
	user := m.Sender

	return UserMap{
		UserID:      int64(user.ID),
		GroupID:     int64(m.Chat.ID),
		DisplayName: user.FirstName + utils.ShouldAddSpace(user) + user.LastName,
		Username:    user.Username,
	}
}
