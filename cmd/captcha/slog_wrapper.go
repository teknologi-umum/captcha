package main

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/allegro/bigcache/v3"
)

// slowWriterWrapper is a wrapper for slog.Logger to implement io.Writer as well as bigcache.Logger
type slogWriterWrapper struct {
	logger *slog.Logger
}

// Printf implements bigcache.Logger.
func (s *slogWriterWrapper) Printf(format string, v ...any) {
	s.logger.Debug(fmt.Sprintf(format, v...))
}

// Write implements io.Writer.
func (s *slogWriterWrapper) Write(p []byte) (n int, err error) {
	s.logger.Debug(string(p))
	return len(p), nil
}

var _ io.Writer = (*slogWriterWrapper)(nil)
var _ bigcache.Logger = (*slogWriterWrapper)(nil)
