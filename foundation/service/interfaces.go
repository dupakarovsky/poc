package foundationservice

import "context"

// Service defines the type of behaviour any concrete type can implement
// in order to...
type Service interface {
	Name() string
	Run() error
	Shutdown(ctx context.Context) error
	Close() error
}
