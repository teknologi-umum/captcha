package reminder_test

import (
	"github.com/teknologi-umum/captcha/reminder"
	"reflect"
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
			input: "me for dinner at 5 pm",
			expect: reminder.Reminder{
				Subject: []string{"me"},
				Time:    time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60)),
				Object:  "dinner",
			},
		},
		{
			name:  "2 person subject",
			input: "@Carl and @Eugene in 2 hours for submit the math assignment",
			expect: reminder.Reminder{
				Subject: []string{"@Carl", "@Eugene"},
				Time:    time.Date(now.Year(), now.Month(), now.Day(), now.Add(time.Hour*2).Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.FixedZone("UTC+7", 7*60*60)),
				Object:  "submit the math assignment",
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := reminder.ParseText(testCase.input)
			if err != nil {
				t.Errorf("[%d] %s", i, err)
			}

			if !reflect.DeepEqual(testCase.expect, got) {
				t.Errorf("[%d] mismatched: (expect) %+v\n(got) %+v", i, testCase.expect, got)
			}
		})
	}
}
