package reminder_test

import (
	"errors"
	"github.com/teknologi-umum/captcha/reminder"
	"strconv"
	"testing"
)

func TestParseClock(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		expectHour   int
		expectMinute int
		expectError  error
	}{
		{
			name:         "happy case 1",
			input:        "00:00",
			expectHour:   0,
			expectMinute: 0,
			expectError:  nil,
		},
		{
			name:         "happy case 2",
			input:        "23:59",
			expectHour:   23,
			expectMinute: 59,
			expectError:  nil,
		},
		{
			name:         "happy case 3",
			input:        "05:05",
			expectHour:   5,
			expectMinute: 5,
			expectError:  nil,
		},
		{
			name:         "happy case 4",
			input:        "20:20",
			expectHour:   20,
			expectMinute: 20,
			expectError:  nil,
		},
		{
			name:         "hour is not a number",
			input:        "abc:00",
			expectHour:   0,
			expectMinute: 0,
			expectError:  strconv.ErrSyntax,
		},
		{
			name:         "minute is not a number",
			input:        "15:abc",
			expectHour:   15,
			expectMinute: 0,
			expectError:  strconv.ErrSyntax,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			hour, minute, err := reminder.ParseClock(testCase.input)
			if hour != testCase.expectHour {
				t.Errorf("expecting hour to be %d, got %d", testCase.expectHour, hour)
			}

			if minute != testCase.expectMinute {
				t.Errorf("expecting minute to be %d, got %d", testCase.expectMinute, minute)
			}

			if !errors.Is(err, testCase.expectError) {
				t.Errorf("expecting error to be %v, got %v", testCase.expectError, err)
			}
		})
	}
}
