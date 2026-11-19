package foundationservice

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// RunAndShutdown will execute the graceful shutdown to each service started
// It now accepts a slice of concrete types that implement the Service interface
func RunAndShutdown(logger *slog.Logger, shutdownTimeout time.Duration, services ...Service) error {

	// guard againts empty slices
	if len(services) == 0 {
		return errors.New("no services provided")
	}

	// capture shutdown signal
	sig := make(chan os.Signal, 1)

	// each goroutine for a service needs to send to the errs chanel. We need
	// as many slots as we have services so the channel is not blocked when a
	// a single send happens.
	errs := make(chan error, len(services))

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// start each of the services (Run) inside their own goroutine
	for i := range services {
		go func() {
			errs <- services[i].Run()
			logger.Info("service goroutine has exited", "service", services[i].Name())
		}()
	}

	// runErr will capture only legitimeg services errors
	var runErr error

	select {
	case runErr = <-errs:
		logger.Error("service run error", "error", runErr)
		// === DONT RETURN HERE ===
		// Otherwise, a failure in any service will skip the shutdown sequece initializaiton

	case s := <-sig:
		logger.Warn("shutdown signal received. beginning shutdown", "signal", s)
	}

	// context for running the Shutdown command
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// We need to shutdown every Service type concurrently
	// ---------------------------------------------------

	var wg sync.WaitGroup
	wg.Add(len(services))
	for i := range services {
		// each service will be terminated inside it's own goroutine
		go func() {
			defer wg.Done()
			//  call each services Shutdown() method
			if err := services[i].Shutdown(ctx); err != nil {
				logger.Error("service failed to shutdown gracefully, forcing close", "service", services[i].Name(), "error", err)

				// Shutdown() faield; call Close()
				if cerr := services[i].Close(); cerr != nil {
					logger.Error("service close also failed , forcing close", "service", services[i].Name(), "error", cerr)
				}
			}
		}()
	}

	// wait for all services to terminate
	wg.Wait()

	logger.Info("shutdown complete")
	return runErr
}
