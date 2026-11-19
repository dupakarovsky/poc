package foundationpostgres

import (
	"time"
)

type SSLMode string

const SSLModeDisabled SSLMode = "disable"
const SSLModeAllow SSLMode = "allow"
const SSLModePrefer SSLMode = "prefer"
const SSLModeRequire SSLMode = "require"
const SSLModeVerifyCA SSLMode = "verify-ca"
const SSLModeVerifyFull SSLMode = "verify-full"

func (s SSLMode) String() string {
	switch s {
	case SSLModeDisabled:
		return "disable"
	case SSLModeRequire:
		return "require"
	case SSLModePrefer:
		return "prefer"
	case SSLModeAllow:
		return "allow"
	case SSLModeVerifyCA:
		return "verify-ca"
	case SSLModeVerifyFull:
		return "verify-full"
	default:
		return "unkown mode"
	}
}

// Config holds every value NewPool needs to build and tune a
// pgxpool.Pool. Deliberately its own type in this package rather than
// folded into cmd/api's config — this package knows nothing about CLI
// flags or env vars, main.go knows nothing about pgxpool internals; main
// builds one of these and hands it in, same dependency direction as
// passing a *slog.Logger into NewCreateReservation.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Schema   string  // used to set search_path so unqualified queries can resolve correctly
	SSLMode  SSLMode // e.g: "disable" | "require | alllow ..."

	// Pool tuning knobs. All left at Go's zero value by default — NewPool
	// only overrides pgxpool's own built-in defaults when a caller
	// actually supplies a non-zero value here (see NewPool's comment for
	// why that distinction matters).
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}
