package analytics

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/telebot.v3"
)

type NullInt64 sql.NullInt64

// GroupMember contains information about a user from a group.
type GroupMember struct {
	GroupID     NullInt64 `json:"group_id,omitempty" db:"group_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Username    string    `json:"username,omitempty" db:"username" redis:"username"`
	DisplayName string    `json:"display_name,omitempty" db:"display_name" redis:"display_name"`
	Counter     int       `json:"counter" db:"counter" redis:"counter"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

// ParseGroupMember converts the tb.Message struct into a GroupMember struct.
func ParseGroupMember(m *tb.Message) GroupMember {
	user := m.Sender

	return GroupMember{
		UserID:      int64(user.ID),
		GroupID:     NullInt64{Int64: m.Chat.ID, Valid: true},
		DisplayName: user.FirstName + utils.ShouldAddSpace(user) + user.LastName,
		Username:    user.Username,
	}
}

func (n *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*n = NullInt64{i.Int64, false}
	} else {
		*n = NullInt64{i.Int64, true}
	}

	return nil
}

func (n *NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Int64)
}

func (n *NullInt64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &n.Int64)
	n.Valid = err == nil
	return err
}

func (n NullInt64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return n.Int64, nil
}
