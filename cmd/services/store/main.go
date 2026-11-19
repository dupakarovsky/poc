package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	foundationconfig "proofofconcept/foundation/config"
	foundationhttp "proofofconcept/foundation/http"
	foundationlogger "proofofconcept/foundation/logger"
	foundationpostgres "proofofconcept/foundation/postgres"
	foundationservice "proofofconcept/foundation/service"
	"proofofconcept/internal/api"
	application "proofofconcept/internal/application/smoke"
	"proofofconcept/internal/repository"
)

func main() {

	//============================================================================
	// CONFIGURATION STRUCT
	//============================================================================
	// generate the service configuraiton file
	cfg, err := foundationconfig.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//============================================================================
	// LOGGER INITIALIZATION
	//============================================================================
	// generate the structure logger for this service. close the log file when main exists
	logger, logFile, err := foundationlogger.New(cfg.Logger.FilePath, cfg.Logger.TermLevel, cfg.Logger.FileLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logFile.Close()

	ctx := context.Background()

	//============================================================================
	// POSTGRES INITIALIZATION
	//============================================================================

	// define the configuration for connecting to the postgres database
	postgresCfg := foundationpostgres.Config{
		Host:            cfg.DB.Host,
		Port:            cfg.DB.Port,
		User:            cfg.DB.User,
		Password:        cfg.DB.Password,
		Database:        cfg.DB.Database,
		Schema:          cfg.DB.Schema,
		SSLMode:         foundationpostgres.SSLMode(cfg.DB.SSLMode),
		MaxConns:        int32(cfg.DB.MaxConns),
		MinConns:        int32(cfg.DB.MinConns),
		MaxConnLifetime: cfg.DB.MaxConnLifetime,
		MaxConnIdleTime: cfg.DB.MaxConnIdleTime,
	}

	// create a new pool of connections
	pool, err := foundationpostgres.NewPool(ctx, postgresCfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger.Info("🗄️ postgres pool ready",
		"host", cfg.DB.Host,
		"port", cfg.DB.Port,
		"user", cfg.DB.User,
		"database", cfg.DB.Database,
	)

	//============================================================================
	// REPOSITORY INITIALIZATION
	//============================================================================
	smokeRepo := repository.NewSmokeRepository(pool)

	//============================================================================
	// APP INITIALIZATION
	//============================================================================
	smokeApp := application.NewSmokeApp(smokeRepo, logger)

	//============================================================================
	// APP INITIALIZATION
	//============================================================================
	smokeAPI := api.NewSmokeAPI(smokeApp)

	//============================================================================
	// ROUTER DEFINITION
	//============================================================================
	router := foundationhttp.NewRouter()

	// apply middleware stack
	router.Use(
		foundationhttp.LoggingMiddleware(logger),
		foundationhttp.RecoveryMiddleware(logger),
		foundationhttp.TimeoutMiddleware(cfg.Server.RequestTimeout),
	)

	// attach Routes
	router.Handle("POST /hello", smokeAPI.Hello)
	router.Handle("POST /smoke/new", smokeAPI.NewSmoke)
	// ---------------------------------------------------------------------------

	// create the handler
	handler := router.Build()

	//============================================================================
	// HTTP SERVER DEFINITION
	//============================================================================
	srv := &http.Server{
		Addr:              cfg.Server.Addr(),
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		Handler:           handler,
	}

	// initialize the http service
	pocHTTPService := foundationservice.NewHTTPService("POC-HTTP", logger, srv)

	// start the service
	if err := foundationservice.RunAndShutdown(logger, cfg.Server.ShutdownTimeout, pocHTTPService); err != nil {
		logger.Error("server existed with error", "error", err)
	}
}
