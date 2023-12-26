package reminder_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/reminder"
)

func TestParseText(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name   string
		input  string
		expect reminder.Reminder
	}{
		{
			name:  "Regular case",
			input: "me for dinner at 17:00",
			expect: reminder.Reminder{
				Subject: []string{"me"},
				Time:    time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60)),
				Object:  "dinner",
			},
		},

		{
			name:  "Time in HH:mm",
			input: "me at 16:23 for assigning task for Jake on Jira",
			expect: reminder.Reminder{
				Subject: []string{"me"},
				Time:    time.Date(now.Year(), now.Month(), now.Day(), 16, 23, 0, 0, time.FixedZone("UTC+7", 7*60*60)),
				Object:  "assigning task for Jake on Jira",
			},
		},
		{
			name:  "Time in X minutes",
			input: "me in 5 minutes for assigning task for Jake on Jira",
			expect: reminder.Reminder{
				Subject: []string{"me"},
				Time:    now.Add(time.Minute * 5).In(time.FixedZone("UTC+7", 7*60*60)).Round(time.Second),
				Object:  "assigning task for Jake on Jira",
			},
		},
		{
			name:  "2 person subject",
			input: "@Carl and @Eugene in 1 hour for submit the math assignment",
			expect: reminder.Reminder{
				Subject: []string{"@Carl", "@Eugene"},
				Time:    now.Add(time.Hour).In(time.FixedZone("UTC+7", 7*60*60)).Round(time.Second),
				Object:  "submit the math assignment",
			},
		},
		{
			name:  "3 person subject",
			input: "me and @Carl and @Eugene to do math homework in 1 minute",
			expect: reminder.Reminder{
				Subject: []string{"me", "@Carl", "@Eugene"},
				Time:    now.Add(time.Minute).In(time.FixedZone("UTC+7", 7*60*60)).Round(time.Second),
				Object:  "do math homework",
			},
		},
		{
			name:  "more than 3 person",
			input: "@Jake and @Jones and @Carl and @August to attend the bachelor party in 3 hours",
			expect: reminder.Reminder{
				Subject: []string{"@Jake", "@Jones", "@Carl"},
				Time:    now.Add(time.Hour * 3).In(time.FixedZone("UTC+7", 7*60*60)).Round(time.Second),
				Object:  "attend the bachelor party",
			},
		},
		{
			name:  "happy case indonesian",
			input: "saya dalam 5 menit untuk mematikan TV",
			expect: reminder.Reminder{
				Subject: []string{"me"},
				Time:    now.Add(time.Minute * 5).In(time.FixedZone("UTC+7", 7*60*60)).Round(time.Second),
				Object:  "mematikan TV",
			},
		},
		{
			name:  "invalid time and object",
			input: "saya in 3 weeks on having a child",
			expect: reminder.Reminder{
				Subject: []string{"me"},
				Time:    time.Time{},
				Object:  "",
			},
		},
	}

	for i, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.expect.Time.Unix() < time.Now().Unix() {
				testCase.expect.Time = testCase.expect.Time.Add(time.Hour * 24)
			}

			ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub())

			got, err := reminder.ParseText(ctx, testCase.input)
			if err != nil {
				t.Errorf("[%d] %s", i, err)
			}

			expectJSON, _ := json.Marshal(testCase.expect)
			gotJSON, _ := json.Marshal(got)

			if !bytes.Equal(expectJSON, gotJSON) {
				t.Errorf("[%d] mismatched: (expect) %v\n(got) %v", i, string(expectJSON), string(gotJSON))
			} else {
				t.Logf("[%d] PASSED", i)
			}
		})
	}

	t.Run("Duration exceeds 24 hours", func(t *testing.T) {
		ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub())

		testInputs := []string{
			"me in 60 hours about something",
			"me in 7200 minutes about something",
		}
		for i, testInput := range testInputs {
			_, err := reminder.ParseText(ctx, testInput)
			if !errors.Is(err, reminder.ErrExceeds24Hours) {
				t.Errorf("[%d] expecting an error of ErrExceeds24Hours, instead got %v", i, err)
			}
		}
	})
}
