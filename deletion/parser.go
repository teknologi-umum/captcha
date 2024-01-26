package deletion

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

func ScanSpace(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, ' '); i >= 0 {
		// We have a full space-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

var clockRegex = regexp.MustCompile("^[0-9]{1,2}:[0-9]{2}(:[0-9]{2})?$")

func ParseDuration(ctx context.Context, s string) (time.Duration, error) {
	span := sentry.StartSpan(ctx, "deletion.parse_duration")
	defer span.Finish()

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(ScanSpace)

	var duration time.Duration = 0
	for scanner.Scan() {
		part := scanner.Text()

		if slices.Contains([]string{"dalam", "pada", "in", "on", "at"}, part) {
			continue
		}

		if value, err := strconv.ParseInt(part, 10, 64); value > 0 && err == nil {
			if duration == 0 {
				duration = time.Duration(value)
				continue
			}
			// don't accept anything if the duration is not empty
			break
		}

		if clockRegex.MatchString(part) {
			if duration != 0 {
				continue
			}

			hour, minute, err := ParseClock(part)
			if err != nil {
				return 0, err
			}

			n := time.Now().Round(time.Second)
			t := time.Date(n.Year(), n.Month(), n.Day(), hour, minute, 0, 0, time.FixedZone("UTC+7", 7*60*60))
			duration = t.Sub(n)
			break
		}

		switch part {
		case "detik", "second", "seconds", "sec", "s":
			duration = duration * time.Second
			break
		case "menit", "m", "mnt", "minute", "minutes":
			duration = duration * time.Minute
			break
		case "jam", "j", "hour", "hours":
			duration = duration * time.Hour
			break
		default:
			continue
		}
	}

	return duration, nil
}

var ErrParseClock = errors.New("parse clock")

func ParseClock(s string) (hour int, minute int, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanRunes)

	var colonMark bool
	var s_hour string
	var s_minute string
	for scanner.Scan() {
		t := scanner.Text()
		if t == ":" {
			if !colonMark {
				colonMark = true
				continue
			}

			break
		}

		if !colonMark {
			s_hour += t
		} else {
			s_minute += t
		}
	}

	hour, err = strconv.Atoi(s_hour)
	if err != nil {
		return
	}

	minute, err = strconv.Atoi(s_minute)
	if err != nil {
		return
	}

	if hour >= 24 {
		err = fmt.Errorf("%w: invalid hour, exceeds 24", ErrParseClock)
		return
	}

	if hour < 0 {
		err = fmt.Errorf("%w: invalid hour, negative number", ErrParseClock)
		return
	}

	if minute >= 60 {
		err = fmt.Errorf("%w: invalid minute, exceeds 60", ErrParseClock)
		return
	}

	if minute < 0 {
		err = fmt.Errorf("%w: invalid minute, negative number", ErrParseClock)
		return
	}

	return
}
