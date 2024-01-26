package deletion_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/teknologi-umum/captcha/deletion"
)

func TestParseDuration(t *testing.T) {
	n := time.Now().Round(time.Second)
	testCases := []struct {
		input  string
		expect time.Duration
	}{
		{
			input:  "in 1 minute",
			expect: time.Minute,
		},
		{
			input:  "dalam 1 menit",
			expect: time.Minute,
		},
		{
			input:  "dalam 30 detik",
			expect: time.Second * 30,
		},
		{
			input:  "in 321 hour",
			expect: time.Hour * 321,
		},
		{
			input:  "in 1 minute 10 second",
			expect: time.Minute,
		},
		{
			input:  "in 11:30",
			expect: time.Date(n.Year(), n.Month(), n.Day(), 11, 30, 0, 0, time.FixedZone("UTC+7", 7*60*60)).Sub(n),
		},
	}

	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := context.Background()
			got, err := deletion.ParseDuration(ctx, testCase.input)
			if err != nil {
				t.Errorf("unexpected error: %s", err.Error())
			}

			if got != testCase.expect {
				t.Errorf("expecting %v, got %v", testCase.expect, got)
			}
		})
	}
}
