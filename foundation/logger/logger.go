package foundationlogger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// multiHandler fans out every log record to multiple underlying handlers.
// This lets us send the same log record to two different destinations, each
// formatted for a different audience: a human at a terminal, and a machine
// parsing a JSON file later.
type multiHandler struct {
	handlers []slog.Handler
}

// newMultiHandler initializes multiHandle struct type
func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{
		handlers: handlers,
	}
}

// Enabled reports whether the record should be handled at all. We check
// every underlying handler and enable if *any* of them wants the record —
// each handler can still separately decide to filter it out in Handle.
func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle passes the record to every underlying handler in turn. If one
// fails, we still attempt the rest rather than aborting early — a failure
// writing to the log file shouldn't silence console output, and vice versa.
func (m *multiHandler) Handle(ctx context.Context, rec slog.Record) error {
	var firstErr error

	// pass the record to every handler
	for _, h := range m.handlers {

		// ignore this handler if not enabled
		if !h.Enabled(ctx, rec.Level) {
			continue
		}

		// Clone the record per handler: slog.Record's Attrs are consumed
		// via an internal iterator, so passing the same record to multiple
		// handlers requires each to get its own clone.
		if err := h.Handle(ctx, rec.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// WithAttrs adds in customs attributes to be exibit by each of the loggers
// in our custom type
func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {

	// slice of handlers to be created, where each now has
	// the custom attributes
	newHandlers := make([]slog.Handler, len(m.handlers))

	// add in the custom attributes to each of the handlers
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}

	// return our custom type where each handler uses our
	// the recevied attributes
	return newMultiHandler(newHandlers...)
}

// WithGroup will allow the custom type to implement the slog.Handlers interface
func (m *multiHandler) WithGroup(name string) slog.Handler {

	// slice of slog.handler, where each handler now has
	// give group
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}

	// return a new instance of our custom type
	return newMultiHandler(newHandlers...)
}

// New builds a *slog.Logger that writes human-readable text to the console
// and structured JSON to logFilePath, both at DEBUG level. It returns the
// logger, the open file handle (the caller is responsible for closing it,
// typically via defer in main), and any error opening the file.
func New(logFilePath, termLevel, fileLevel string) (*slog.Logger, *os.File, error) {

	// create a new handler to write exclusively to the terminal
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(termLevel), AddSource: true, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

			// if the current attribute is the "time" attribue
			if a.Key == slog.TimeKey {
				// replace it with the RFC3339Nano format
				if source, ok := a.Value.Any().(time.Time); ok {
					v := source.Format(time.RFC3339Nano)
					return slog.Attr{Key: "time", Value: slog.StringValue(v)}
				}
			}

			// if the current attribute is the source key attribute, which is used
			// to output the source file and line
			if a.Key == slog.SourceKey {
				// get the value from the source key
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// replace the attribute's value with [󰡱 main] main.go:5
					v := fmt.Sprintf("[%s:%d]", filepath.Base(source.File), source.Line)
					// return a new attribute
					return slog.Attr{Key: "source", Value: slog.StringValue(v)}

				}
			}

			return a
		},
	})

	// if no output file is passed in, we create a standard slog handler that
	// outputs only to stdout. This is be needed when operating in a container
	// so the logs go to docker compose logs.
	if logFilePath == "" {
		return slog.New(consoleHandler), nil, nil
	}

	// open|create the log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}

	// create a new handler to write exclusively to file output
	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: parseLevel(fileLevel), AddSource: true})

	// create a multiHandler
	multiHandler := newMultiHandler(consoleHandler, fileHandler)

	// create a new logger using our custom type
	logger := slog.New(multiHandler)

	return logger, logFile, nil
}

// NewWithAttrs constructs custom slog.Logger with custom attributs
func NewWithAttrs(logFilePath, termLevel, fileLevel string, args ...any) (*slog.Logger, *os.File, error) {

	// initializes a new slog handler given a log file, the terminal level of
	// serverity (defaults to DEBUG) and file level or sevrity (defaults to INFO)
	handler, file, err := New(logFilePath, termLevel, fileLevel)
	if err != nil {
		return nil, nil, err
	}

	// pass the additional argus to the handler
	newHandler := handler.With(args...)

	// return it slog.Logger as well as the log file
	return newHandler, file, err
}

// NewWithAttrs constructs custom slog.Logger with custom attributs
func NewWithGroup(logFilePath, termLevel, fileLevel, group string) (*slog.Logger, *os.File, error) {

	handler, file, err := New(logFilePath, termLevel, fileLevel)
	if err != nil {
		return nil, nil, err
	}

	newHandler := handler.WithGroup(group)

	return newHandler, file, err
}

//-----------------------------------------------------------------------------

// parseLevel converts a level string ("debug", "info", "warn", "error",
// case-insensitive) into the corresponding slog.Level. Falls back to
// slog.LevelInfo on anything unrecognized — a typo'd flag value
// (--log-term-level=verbose, say) should degrade to a safe, moderate
// default rather than crash the server at startup over a logging
// preference.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
