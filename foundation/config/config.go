package foundationconfig

import (
	"fmt"
	"net"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds the three distincts configuration structs: one for
// the server, one for the logger and one for the database
type Config struct {
	Server ServerConfig
	Logger LoggerConfig
	DB     DBConfig
}

// ServerConfig defines the configuration parameters for the server
// using the envcofig package
type ServerConfig struct {
	Host              string        `envconfig:"SRV_HOST" default:""`
	Port              string        `envconfig:"SRV_PORT" default:"8080"`
	ReadHeaderTimeout time.Duration `envconfig:"SRV_READ_HEADER_TIMEOUT" default:"5s"`
	WriteTimeout      time.Duration `envconfig:"SRV_WRITE_TIMEOUT" default:"13s"`
	IdleTimeout       time.Duration `envconfig:"SRV_IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout   time.Duration `envconfig:"SRV_SHUTDOWN_TIMEOUT" default:"10s"`
	RequestTimeout    time.Duration `envconfig:"SRV_REQUEST_TIMEOUT" default:"12s"`
}

// Addr returns the joinded string of the Host and Port fields, which is
// compatible with IPV4 and IPV6
func (s ServerConfig) Addr() string {
	return net.JoinHostPort(s.Host, s.Port)
}

// Logger config defines configuration params for the log's level in the terminal
// and in the .log file
type LoggerConfig struct {
	FilePath  string `envconfig:"LOG_FILE_PATH" default:""`
	FileLevel string `envconfig:"LOG_FILE_LEVEL" default:"info"`
	TermLevel string `envconfig:"LOG_TERM_LEVEL" default:"debug"`
}

// DBConfig defines configuration paramters for the postgres database
type DBConfig struct {
	Host            string        `envconfig:"DB_HOST" default:"localhost"`
	Port            int           `envconfig:"DB_PORT" default:"5432"`
	User            string        `envconfig:"DB_USER"`
	Password        string        `envconfig:"DB_PASSWORD"`
	Database        string        `envconfig:"DB_DATABASE" default:""`
	Schema          string        `envconfig:"DB_SCHEMA"`
	SSLMode         string        `envconfig:"DB_SSL_MODE" default:"require"`
	MaxConns        int           `envconfig:"DB_MAX_CONNS" default:"15"`
	MinConns        int           `envconfig:"DB_MIN_CONNS" default:"5"`
	MaxConnLifetime time.Duration `envconfig:"DB_MAX_CONNS_LIFETIME" default:"1h"`
	MaxConnIdleTime time.Duration `envconfig:"DB_MAX_CONNS_IDLE_TIME" default:"30m"`
}

// Load reads the .env file and injects the key/value pairs using os.Setenv
// Missing .env file outputs a print and is ignored. In production we set environmental
// variables and don't rely on .env file
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("no .env file found, relying on real environmental variables")
	}

	// envconfig.Process is the population step. First agument is the prefix (passing "APP",
	// would result in "APP_SRV_PORT" insted of "SRV_PORT"). It walks the Config fields via
	// reflection, matching each tag to an environmental variable.
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("processing config: %w", err)
	}

	return cfg, nil
}
