package main

// Copied from samber/slog-multi
// https://github.com/samber/slog-multi/blob/e7e598c398ffadc7ab7e5b73f545115e07aadc73/multi.go
//
// MIT License
//
// Copyright (c) 2023 Samuel Berthe
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
)

func try(callback func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("unexpected error: %+v", r)
			}
		}
	}()

	err = callback()

	return
}

var _ slog.Handler = (*FanoutHandler)(nil)

type FanoutHandler struct {
	handlers []slog.Handler
}

// SlogFanout distributes records to multiple slog.SlogSentryBreadcrumb in parallel
func SlogFanout(handlers ...slog.Handler) slog.Handler {
	return &FanoutHandler{
		handlers: handlers,
	}
}

// Implements slog.SlogSentryBreadcrumb
func (h *FanoutHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// Implements slog.SlogSentryBreadcrumb
func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := try(func() error {
				return h.handlers[i].Handle(ctx, r.Clone())
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	// If errs is empty, or contains only nil errors, this returns nil
	return errors.Join(errs...)
}

// Implements slog.SlogSentryBreadcrumb
func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var handlers []slog.Handler
	for i := range h.handlers {
		handlers = append(handlers, h.handlers[i].WithAttrs(slices.Clone(attrs)))
	}

	return SlogFanout(handlers...)
}

// Implements slog.SlogSentryBreadcrumb
func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	var handlers []slog.Handler
	for i := range h.handlers {
		handlers = append(handlers, h.handlers[i].WithGroup(name))
	}

	return SlogFanout(handlers...)
}
