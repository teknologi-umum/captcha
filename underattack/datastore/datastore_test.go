package datastore_test

import (
	"os"
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestMain(m *testing.M) {
	_ = sentry.Init(sentry.ClientOptions{})

	os.Exit(m.Run())
}
