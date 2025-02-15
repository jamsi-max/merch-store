#!/bin/sh

source ./.env

echo "Waiting for PostgreSQL to start..."
until nc -z db 5432; do
  sleep 1
done
echo "PostgreSQL started"

echo "Running migrations..."
goose -dir=./migrations postgres "user=$DB_USER password=$DB_PASS dbname=$DB_NAME host=db sslmode=disable" up

echo "Starting app..."
exec ./merch-store