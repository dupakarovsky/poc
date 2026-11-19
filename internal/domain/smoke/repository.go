package domainsmoke

import (
	"context"
	"errors"
)

var ErrSmokeNotFound = errors.New("smoke not found")

// CustomerRepository defines the behaviour any concrete type should have
// in order to communicate with any database adapter
type SmokeRepository interface {
	//-- write
	Create(ctx context.Context, description string) (*Smoke, error)
	Update(ctx context.Context, smoke *Smoke) error
	FindByIDForUpdate(ctx context.Context, id string) (*Smoke, error)
	WithLock(ctx context.Context, fn func(ctx context.Context) error) error
	//-- read
	FindByID(ctx context.Context, id string) (*Smoke, error)
}
