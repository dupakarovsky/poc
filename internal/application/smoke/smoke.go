package application

import (
	"context"
	"fmt"
	"log/slog"
	domainsmoke "proofofconcept/internal/domain/smoke"
)

// RegisterCustomer is a type that can create a Customer type, communidate
// with the repository and publish the event to the stream
type SmokeApp struct {
	smokeRepo domainsmoke.SmokeRepository
	logger    *slog.Logger
}

// NewRegisterCustomer constructs a new RegisterCustomer type
func NewSmokeApp(smokeRepo domainsmoke.SmokeRepository, logger *slog.Logger) *SmokeApp {
	return &SmokeApp{
		smokeRepo: smokeRepo,
		logger:    logger,
	}
}

// Create will add a new smoke to the database
func (app *SmokeApp) Create(ctx context.Context, smoke *domainsmoke.Smoke) (*domainsmoke.Smoke, error) {

	// save to the repository
	smoke, err := app.smokeRepo.Create(ctx, smoke.Description)
	if err != nil {
		return nil, fmt.Errorf("creating smoke: %w", err)
	}

	// saved
	// --------------------------------

	return smoke, nil
}

// TODO: Not yet implemented
//--------------------------

// Update will perform the update on the smoke on the database
func (app *SmokeApp) Update(ctx context.Context, smoke *domainsmoke.Smoke) error {
	return nil
}

func (app *SmokeApp) FindByIDForUpdate(ctx context.Context, id string) (*domainsmoke.Smoke, error) {
	return &domainsmoke.Smoke{}, nil
}

func (app *SmokeApp) WithLock(ctx context.Context, fn func(ctx context.Context) error) error {
	return nil
}

func (app *SmokeApp) FindByID(ctx context.Context, id string) (*domainsmoke.Smoke, error) {
	return &domainsmoke.Smoke{}, nil
}
