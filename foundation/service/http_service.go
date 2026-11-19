package foundationservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// httpService is a concrete type the implements the Service interface and
// will be used to start a http server
type httpService struct {
	name   string
	logger *slog.Logger
	srv    *http.Server
}

// NewHTTPService will create a new httpService concrete type which has it's own
// server logger and service name. This type's server will go through the Shutdown
// process.
func NewHTTPService(name string, logger *slog.Logger, srv *http.Server) Service {

	return &httpService{
		name:   name,
		logger: logger,
		srv:    srv,
	}
}

// Name returns the name of the current service
func (s *httpService) Name() string {
	return s.name
}

// Run will starts listening to the server and blocks wainting for errors or
// a shutdown
func (s *httpService) Run() error {

	message := fmt.Sprintf("🚀 starting %s", s.Name())

	s.logger.Info(message,
		"srv-addr", s.srv.Addr,
		"srv-read-header-timeout", s.srv.ReadHeaderTimeout.String(),
		"srv-write-timeout", s.srv.WriteTimeout.String(),
		"srv-idle-timeout", s.srv.IdleTimeout.String(),
	)
	// start listening; ErrServerClosed means the server closed correctly
	// and no error needs to be reported
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown is just a simple passthrough to the http.Server Shutdown function
// It will try to perform a Graceful shutdown
func (s *httpService) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

// Shutdown is just a simple passthrought to the http.Server Close() function
// It'll abruptly close all http connections.
func (s *httpService) Close() error {
	return s.srv.Close()
}
