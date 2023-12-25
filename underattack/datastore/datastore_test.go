package datastore_test

import (
	"github.com/getsentry/sentry-go"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_ = sentry.Init(sentry.ClientOptions{})

	os.Exit(m.Run())
}
