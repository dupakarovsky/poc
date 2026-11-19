package domainsmoke

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrDescriptionIsBlank   = errors.New("description cannot be blank")
	ErrDescriptionUnchanged = errors.New("description did not change")
	ErrAlreadyEnabled       = errors.New("smoke is already enabled")
	ErrAlreadyDisabled      = errors.New("smoke is already disabled")
)

// Smoke defines a fake domain to test the application
type Smoke struct {
	ID          int64
	Description string
	Enabled     bool
	CreateAt    time.Time
}

// NewSmoke instantiates a new type
func NewSmoke(description string) (*Smoke, error) {

	// remove spaces in front and and the end
	description = strings.TrimSpace(description)

	// guard against blank name
	if description == "" {
		return nil, ErrDescriptionIsBlank
	}

	// instantiate a new Customer
	smoke := &Smoke{
		Description: description,
		Enabled:     false,
		CreateAt:    time.Now(),
	}

	return smoke, nil
}

// ChangeSmsNumber updates the Customer number, returning an CustomerSmsChangedEvent
func (s *Smoke) UpdateDescription(description string) error {

	// remove leading and trailling spaces
	description = strings.TrimSpace(description)

	// guard against blank numbers
	if description == "" {
		return ErrDescriptionIsBlank
	}

	// guard against the same number
	if description == s.Description {
		return ErrDescriptionUnchanged
	}

	// passed the checks; can be updated
	// ---------------------------------

	// upadate the sms number
	s.Description = description

	return nil
}

// Enable will enabled a disabled Customer
func (s *Smoke) Enable() error {

	if s.Enabled {
		return ErrAlreadyEnabled
	}

	// passed; can now be enabled
	// -----------------------------

	// update the property
	s.Enabled = true

	return nil
}

// Disable will disabled an enabled  Customer
func (s *Smoke) Disable() error {

	// verify if it is already disabled
	if !s.Enabled {
		return ErrAlreadyDisabled
	}

	// can be disabled
	//------------------

	// update the property
	s.Enabled = false

	return nil
}

// ReconstituteSmoke will reconstruct an Smoke type from the primitives
// parameters it's given
func ReconstituteSmoke(id int64, description string, enabled bool, createdAt time.Time) *Smoke {
	return &Smoke{
		ID:          id,
		Description: description,
		Enabled:     enabled,
		CreateAt:    createdAt,
	}
}
