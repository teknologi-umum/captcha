package analytics_test

import (
	"testing"

	"teknologi-umum-captcha/analytics"

	tb "gopkg.in/telebot.v3"
)

func TestParseGroupMember(t *testing.T) {
	message := &tb.Message{
		Chat: &tb.Chat{
			ID:   123456789,
			Type: tb.ChatGroup,
		},
		Sender: &tb.User{
			ID:        1,
			FirstName: "Reinaldy",
			LastName:  "Reinaldy",
			Username:  "reinaldy",
		},
	}

	userMap := analytics.ParseGroupMember(message)
	if userMap.UserID != 1 {
		t.Errorf("UserID should be 1, got: %d", userMap.UserID)
	}
	if userMap.DisplayName != "Reinaldy Reinaldy" {
		t.Errorf("DisplayName should be Reinaldy Reinaldy, got: %s", userMap.DisplayName)
	}
	if userMap.Username != "reinaldy" {
		t.Errorf("Username should be reinaldy, got: %s", userMap.Username)
	}
	if userMap.GroupID.Int64 != 123456789 {
		t.Errorf("GroupID should be 123456789, got: %d", userMap.GroupID.Int64)
	}
}

func TestNullInt64(t *testing.T) {
	t.Run("get valid value", func(t *testing.T) {
		n := analytics.NullInt64{
			Int64: 30000,
			Valid: true,
		}

		v, err := n.Value()
		if err != nil {
			t.Error(err)
		}

		val, ok := v.(int64)
		if !ok {
			t.Error(err)
		}

		if val != 30000 {
			t.Errorf("Value should be 30000, got: %d", val)
		}
	})

	t.Run("get invalid value", func(t *testing.T) {
		n := analytics.NullInt64{
			Int64: 0,
			Valid: false,
		}

		v, _ := n.Value()
		if v != nil {
			t.Errorf("Value should be nil, got: %v", v)
		}
	})

	t.Run("json marshal", func(t *testing.T) {
		n1 := &analytics.NullInt64{
			Int64: 30_000,
			Valid: true,
		}

		b, err := n1.MarshalJSON()
		if err != nil {
			t.Error(err)
		}

		if string(b) != "30000" {
			t.Errorf("JSON should be 30000, got: %s", string(b))
		}

		n2 := &analytics.NullInt64{
			Int64: 0,
			Valid: false,
		}

		b, err = n2.MarshalJSON()
		if err != nil {
			t.Error(err)
		}

		if string(b) != "null" {
			t.Errorf("JSON should be null, got: %s", string(b))
		}
	})

	t.Run("json unmarshal", func(t *testing.T) {
		n1 := &analytics.NullInt64{}

		err := n1.UnmarshalJSON([]byte("30000"))
		if err != nil {
			t.Error(err)
		}

		if n1.Int64 != 30_000 {
			t.Errorf("Value should be 30_000, got: %d", n1.Int64)
		}

		if !n1.Valid {
			t.Error("Value should be valid")
		}

		n2 := &analytics.NullInt64{}

		err = n2.UnmarshalJSON([]byte("null"))
		if err != nil {
			t.Error(err)
		}

		if n2.Int64 != 0 {
			t.Errorf("Value should be 0, got: %d", n2.Int64)
		}

		if !n2.Valid {
			t.Error("Value should be valid")
		}

		// test invalid value

		n3 := &analytics.NullInt64{}

		err = n3.UnmarshalJSON([]byte("invalid"))
		if err == nil {
			t.Error("Should be error")
		}

		if n3.Valid {
			t.Error("Value should be invalid")
		}
	})
}
