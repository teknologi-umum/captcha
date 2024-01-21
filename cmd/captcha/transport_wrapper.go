package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
)

type SentryTransportWrapper struct {
	OriginalTransport *http.Transport
}

func (s *SentryTransportWrapper) RoundTrip(request *http.Request) (*http.Response, error) {
	ignoredMethods := []string{"getme", "logout", "close", "getupdates"}
	if slices.Contains(ignoredMethods, strings.ToLower(request.URL.Path)) {
		return s.OriginalTransport.RoundTrip(request)
	}

	// Start Sentry trace
	ctx := request.Context()
	var cleanRequestURL string
	pathFragments := strings.Split(request.URL.Path, "/")
	if len(pathFragments) >= 2 {
		if pathFragments[0] == "" {
			pathFragments[1] = "bot[Filtered]"
		} else {
			pathFragments[0] = "bot[Filtered]"
		}

		cleanRequestURL = strings.Join(pathFragments, "/")
	}

	span := sentry.StartSpan(ctx, "http.client", sentry.WithTransactionName(fmt.Sprintf("%s %s", request.Method, cleanRequestURL)))
	defer span.Finish()

	span.SetData("http.query", request.URL.Query().Encode())
	span.SetData("http.fragment", request.URL.Fragment)
	span.SetData("http.request.method", request.Method)

	response, err := s.OriginalTransport.RoundTrip(request)

	if response != nil {
		span.Status = sentry.HTTPtoSpanStatus(response.StatusCode)
		span.SetData("http.response.status_code", response.Status)
		span.SetData("http.response_content_length", strconv.FormatInt(response.ContentLength, 10))
	}

	return response, err
}
