package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type envelope map[string]any

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	const maxBytes = 16 << 10

	// limits the request body to 16kb only
	body := http.MaxBytesReader(w, r.Body, maxBytes)

	// start a new decoder
	decoder := json.NewDecoder(body)

	// restrict unknow field
	decoder.DisallowUnknownFields()

	// Decode the request body into the target destination.
	err := decoder.Decode(dst)
	if err != nil {
		// If there is an error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// invalid fields
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", field)

		// body exceeds limit
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body size exceeds max limit")

		// malfored json
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		//  mismatched type
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// nil pointer
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// go bug: bad json. eg: {'filed', value, 'field: value}
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// empty body
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// For anything else, return the error message as-is.
		default:
			return err
		}
	}

	// verify if second body is comming
	// if it's not the end of the file, then there's a second body
	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("body must contain only a single JSON value")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {

	// marshal the data with indentation
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling json: %w", err)
	}

	// add a line break
	js = append(js, '\n')

	maps.Copy(w.Header(), headers)

	// add any special http headers and write the status code.
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(js); err != nil {
		return fmt.Errorf("error writing json: %w", err)
	}
	return nil
}

func readCSV(qs url.Values, key string, defaultVals []string) []string {

	// get the value associated with the query string key
	csv := qs.Get(key)

	// in not passed, return a default value
	if csv == "" {
		return defaultVals
	}

	return strings.Split(csv, ",")
}

// background will launch a goroutine in the background with a function that will execute concurrently
func background(fn func(), status string, logger *slog.Logger) {

	wg := sync.WaitGroup{}

	// add 1 wait group for the goroutine
	wg.Add(1)

	// launch a goroutine
	go func() {
		defer wg.Done()

		logger.Info("  lauching goroutine | background goroutine", "status", status)

		// execute when goroutine returns to prevent panics
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovery", "error", err)
			}
		}()

		// execute the function
		fn()

		logger.Info("  terminating goroutine | background goroutine", "status", status)
	}()
}
