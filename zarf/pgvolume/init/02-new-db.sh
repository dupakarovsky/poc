#!/bin/bash

set -e

psql -v ON_ERROR_STOP=1 --username="$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE smoke (
        id SERIAL PRIMARY KEY,
        description TEXT NOT NULL,
        enable BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
EOSQL
