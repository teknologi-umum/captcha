package main

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/allegro/bigcache/v3"
)

type slogWriterWrapper struct {
	logger *slog.Logger
}

// Printf implements bigcache.Logger.
func (s *slogWriterWrapper) Printf(format string, v ...interface{}) {
	s.logger.Debug(fmt.Sprintf(format, v...))
}

// Write implements io.Writer.
func (s *slogWriterWrapper) Write(p []byte) (n int, err error) {
	s.logger.Debug(string(p))
	return len(p), nil
}

var _ io.Writer = (*slogWriterWrapper)(nil)
var _ bigcache.Logger = (*slogWriterWrapper)(nil)
