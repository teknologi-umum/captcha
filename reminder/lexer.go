package reminder

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

type SentenceElement uint8

const (
	None SentenceElement = iota
	Subject
	Time
	Verb
	TimePreposition
	VerbPreposition
	Conjunction
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

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

var verbPreposition = []string{"for", "untuk", "buat", "to", "about", "tentang", "soal", "regarding"}
var timePreposition = []string{"at", "in", "on", "di", "jam", "pada", "dalam"}
var conjunction = []string{"and", "or", "dan", "&", "atau"}
var validSubjects = []string{"me", "aku", "saya", "gw", "gua", "gue", "gweh"}

var clockRegex = regexp.MustCompile("[0-9]{1,2}:[0-9]{2}")

func ParseText(ctx context.Context, text string) (Reminder, error) {
	span := sentry.StartSpan(ctx, "reminder.parse_text")
	defer span.Finish()

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(ScanSpace)

	var reminder Reminder
	var lastPartCategory SentenceElement = None
	var expectedNextPartCategory SentenceElement = None
	var partialTimeString string
	var appliedVerbPreposition bool
	for scanner.Scan() {
		part := scanner.Text()
		// subject = 0: me, @mention, saya, gw, aku
		// predicate = for, in, at, on

		if len(reminder.Subject) == 0 || expectedNextPartCategory == Subject {
			//ValidateSubject:
			if len(reminder.Subject) == 3 {
				lastPartCategory = Subject
				expectedNextPartCategory = None
				continue
			}

			// does this part contain valid subject?
			if slices.Contains(validSubjects, strings.ToLower(part)) {
				// it is subject indeed
				reminder.Subject = append(reminder.Subject, "me")
				lastPartCategory = Subject
				expectedNextPartCategory = None
				continue
			}

			// check if it starts with '@'
			if strings.HasPrefix(part, "@") {
				// it is subject, yeah
				reminder.Subject = append(reminder.Subject, part)
				lastPartCategory = Subject
				expectedNextPartCategory = None
				continue
			}
		}

		// check for conjunction in parts
		if slices.Contains(conjunction, part) {
			lastPartCategory = Conjunction
			if reminder.Object == "" && reminder.Time.IsZero() {
				expectedNextPartCategory = Subject
			}
			continue
		}

		// check for preposition on time
		if slices.Contains(timePreposition, part) && reminder.Time.IsZero() {
			lastPartCategory = TimePreposition
			expectedNextPartCategory = Time
			continue
		}

		if lastPartCategory == TimePreposition && expectedNextPartCategory == Time {
			// check if it's matches with clock regex
			if clockRegex.MatchString(part) {
				// parse clock
				hour, minute, err := ParseClock(part)
				if err != nil {
					return Reminder{}, fmt.Errorf("parsing clock: %w", err)
				}

				now := time.Now()
				reminder.Time = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.FixedZone("UTC+7", 7*60*60))
				lastPartCategory = Time
				expectedNextPartCategory = Verb
				continue
			}

			// if partialTimeString is empty, check if the current part is number
			if partialTimeString == "" && isNumber(part) {
				partialTimeString = part
				continue
			} else {
				// partial string is not empty
				now := time.Now()

				// switch case is faster rather than doing slices.Contains like this
				//if slices.Contains(timeDuration, part) {}
				switch strings.ToLower(part) {
				case "minute", "minutes", "menit":
					partialTimeString += "m"
					duration, err := time.ParseDuration(partialTimeString)
					if err != nil {
						return reminder, fmt.Errorf("invalid duration: %w", err)
					}

					// must not exceed 24 hours
					if duration > time.Duration(24)*time.Hour {
						return reminder, ErrExceeds24Hours
					}

					reminder.Time = now.Add(duration).
						In(time.FixedZone("UTC+7", 7*60*60)).
						Round(time.Second)
					lastPartCategory = Time
					expectedNextPartCategory = Verb
					continue
				case "hour", "hours", "jam":
					partialTimeString += "h"
					duration, err := time.ParseDuration(partialTimeString)
					if err != nil {
						return reminder, fmt.Errorf("invalid duration: %w", err)
					}

					// must not exceed 24 hours
					if duration > time.Duration(24)*time.Hour {
						return reminder, ErrExceeds24Hours
					}

					reminder.Time = now.Add(duration).
						In(time.FixedZone("UTC+7", 7*60*60)).
						Round(time.Second)
					lastPartCategory = Time
					expectedNextPartCategory = Verb
					continue
				}
			}
		}

		if expectedNextPartCategory == None || expectedNextPartCategory == Verb {
			if slices.Contains(verbPreposition, part) {
				if !appliedVerbPreposition {
					appliedVerbPreposition = true
					lastPartCategory = VerbPreposition
					continue
				}
			}

			reminder.Object += part + " "
		}
	}

	reminder.Object = strings.TrimSpace(reminder.Object)

	// validate for time, it should be later than now.
	// if it's not, then we should add the day by 1?
	if reminder.Time.Unix() < time.Now().Unix() {
		reminder.Time = reminder.Time.Add(time.Hour * 24)
	}

	return reminder, nil
}
