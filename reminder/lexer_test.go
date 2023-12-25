package reminder_test

import (
	"bytes"
	"encoding/json"
	"github.com/teknologi-umum/captcha/reminder"
	"testing"
	"time"
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
				Time:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+5, now.Second(), now.Nanosecond(), time.FixedZone("UTC+7", 7*60*60)).Round(time.Second),
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
	}

	for i, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.expect.Time.Unix() < time.Now().Unix() {
				testCase.expect.Time = testCase.expect.Time.Add(time.Hour * 24)
			}

			got, err := reminder.ParseText(testCase.input)
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
}
