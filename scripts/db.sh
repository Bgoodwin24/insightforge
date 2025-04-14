#!/bin/bash

# Source the environment variables
if [ -f .env ]; then
    source .env
fi

# Connect to PostgreSQL using environment variables
PGPASSWORD=$DB_PASSWORD psql -U $DB_USER -h $DB_HOST -p $DB_PORT -d $DB_NAME
