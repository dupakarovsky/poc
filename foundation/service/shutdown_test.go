package foundationservice_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	foundationservice "proofofconcept/foundation/service"
	"testing"
	"time"
)

// fakeFailingService implements the Service interface so it can
// run as a service in a the shutdown.Run() function
type fakeFailingService struct {
	delay time.Duration
}

func (s *fakeFailingService) Name() string {
	return "MOCKTEST"
}
func (s *fakeFailingService) Run() error {
	time.Sleep(s.delay)
	return errors.New("simulated failure")
}
func (s *fakeFailingService) Shutdown(ctx context.Context) error {
	return nil
}
func (s *fakeFailingService) Close() error {
	return nil
}

func TestRun_WaitsForInFlightRequestBeforeReturning(t *testing.T) {

	const addr = "127.0.0.1:18080"
	const handlerDelay = 400 * time.Millisecond
	const triggerDelay = 100 * time.Millisecond

	// create a server mux
	mux := http.NewServeMux()

	// a slow handler to be called
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// will sleep for 400 Millisecond before responding
		time.Sleep(handlerDelay)
		w.WriteHeader(http.StatusOK)
	})

	// create a new server
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// define a new logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// instantiate a new httpService (the stable service)
	httpService := foundationservice.NewHTTPService("SLOW_HTTP", logger, srv)

	// instantiate a fakeFailingService. This service will fail after 100 Millisecond
	// and will trigger a shutdown process for the RunAndShutdown function while the
	// route /slow is still responding
	trigger := &fakeFailingService{delay: triggerDelay}

	// channel will receive the time the goroutine that httpService
	// is running inside has exited
	runReturned := make(chan time.Time, 1)

	// Run both service inside a goroutine.
	go func() {
		_ = foundationservice.RunAndShutdown(logger, 5*time.Second, httpService, trigger)
		runReturned <- time.Now()
	}()

	// wait for the http server to be online before continuing
	time.Sleep(50 * time.Millisecond)

	// Server should be online now
	// we need to make an http Request to the /slow route
	// ------------------------------------------------------

	type requestResult struct {
		err      error
		status   int
		finished time.Time
	}

	requestDone := make(chan requestResult, 1)

	// make the HTTP request to "/slow" route from a request goroutine
	go func() {
		// make the request and close the response body
		res, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			// send the error to the main goroutine and exit this one
			requestDone <- requestResult{err: err, finished: time.Now()}
			return
		}
		defer res.Body.Close()

		// response was successful.
		// Send the results to the main gououtine
		// --------------------------------------

		// send result data to the chanel and exit
		requestDone <- requestResult{status: res.StatusCode, finished: time.Now()}
	}()

	// capture the requestResult type that goroutine has returned
	result := <-requestDone
	runDoneAt := <-runReturned

	if result.err != nil {
		t.Fatalf("in-flight request did not complete cleanly (%v) — shutdown likely cut it off mid-flight", result.err)
	}
	if result.status != http.StatusOK {
		t.Fatalf("expected 200, got %d", result.status)
	}
	if runDoneAt.Before(result.finished) {
		t.Fatalf("Run() returned at %v, before the in-flight request finished at %v — shutdown did not wait for in-flight work",
			runDoneAt, result.finished)
	}

	t.Logf("Run() returned %v after the in-flight request finished — graceful wait confirmed", runDoneAt.Sub(result.finished))
}
