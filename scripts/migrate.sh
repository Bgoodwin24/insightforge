#!/bin/bash
# Usage: ./scripts/migrate.sh up|down
# Set DB_USER, DB_PASSWORD, DB_NAME environment variables before running

# Source environment variables if .env file exists
if [ -f .env ]; then
    source .env
fi

# Check if required environment variables are set
if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_NAME" ] || [ -z "$DB_SSLMODE" ]; then
  echo "Error: Missing required environment variables"
  echo "Please set: DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME, DB_SSLMODE"
  exit 1
fi

DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
COMMAND=$1

goose -dir migrations postgres "$DB_URL" $COMMAND
