## ============================================================================
## FOR DEPLOYMENT
## ============================================================================

## ----------------------------------------------------------------------------
## --- Containers -- 
## ----------------------------------------------------------------------------
COMPOSE := docker compose

.PHONY:build
build:
	$(COMPOSE) build

.PHONY:up
up:
	$(COMPOSE) up -d

.PHONY:down 
down:
	$(COMPOSE) down


# logs of services
.PHONY:logs 
logs:
	$(COMPOSE) logs

.PHONY:kill
kill/stores:
	$(COMPOSE) kill -s SIGTERM poc

.PHONY:ps
ps:
	$(COMPOSE) ps

.PHONY:exec/pg
exec/pg:
	$(COMPOSE) exec postgres psql -h 127.0.0.1 -U edasuper -d edadb

## ----------------------------------------------------------------------------
## --- Migrations -- 
## ----------------------------------------------------------------------------
include .postgres.env
export

.PHONY: migrate/poc/up migrate/poc/down 

## --- Stores DB -- 
migrate/poc/up:
	@migrate -path stores/migrations \
		-database "postgres://poc_dev:$(DEV_PASSWORD)@$(MIGRATE_DB_HOST):$(MIGRATE_DB_PORT)/pocdb?sslmode=require" \
		up

migrate/poc/down:
	@migrate -path stores/migrations \
		-database "postgres://poc_dev:$(DEV_PASSWORD)@$(MIGRATE_DB_HOST):$(MIGRATE_DB_PORT)/storesdb?sslmode=require" \
		down 1

## ============================================================================
## FOR DEVELOPMENT ONLY
## ============================================================================

## ----------------------------------------------------------------------------
## --- run stores locally -- 
## ----------------------------------------------------------------------------

.PHONY: poc/run 
poc/run:
	@echo 'Running poc service...'
	@mkdir -p ./poc/bin
	GOOS=linux GOARCH=amd64 go build -o=./bin/stores ./cmd/service/poc
	chmod +x ./bin/poc
	./bin/poc


# connect to the database as the dev
.PHONY: poc/dev
poc/dev:
	psql -h 127.0.0.1 -p 5433 -d pocdb -U poc_dev

# connect to the database as the user
.PHONY: poc/user
poc/user:
	psql -h 127.0.0.1 -p 5433 -d pocdb -U poc_user



