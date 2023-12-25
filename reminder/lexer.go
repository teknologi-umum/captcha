package reminder

import (
	"bufio"
	"bytes"
	"slices"
	"strings"
)

type SentenceElement uint8

const (
	None SentenceElement = iota
	Subject
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

var timePreposition = []string{"at", "in", "on", "di", "jam", "pada"}
var conjunction = []string{"and", "or", "dan", "&", "atau"}
var validSubjects = []string{"me", "aku", "saya", "gw", "gua", "gue", "gweh"}

func ParseText(text string) (Reminder, error) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(ScanSpace)

	var reminder Reminder
	var i = 0
	var lastPartCategory SentenceElement
	var expectedNextPartCategory SentenceElement
	for scanner.Scan() {
		part := scanner.Text()
		// subject = 0: me, @mention, saya, gw, aku
		// predicate = for, in, at, on

		switch expectedNextPartCategory {
		case Subject:
			//goto ValidateSubject
		}

		if len(reminder.Subject) == 0 {
			//ValidateSubject:
			if len(reminder.Subject) == 3 {
				continue
			}

			// does this part contain valid subject?
			if slices.Contains(validSubjects, part) {
				// it is subject indeed
				reminder.Subject = append(reminder.Subject, part)
				lastPartCategory = Subject
				continue
			}

			// check if it starts with '@'
			if strings.HasPrefix(part, "@") {
				// it is subject, yeah
				reminder.Subject = append(reminder.Subject, part)
				lastPartCategory = Subject
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

		i++
	}

	return reminder, nil
}
