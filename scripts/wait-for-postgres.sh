#!/bin/sh
# wait-for-postgres.sh

set -e

cmd="$@"

sleep 1
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -p $DB_PORT $DB_NAME -c '\q'; do
  echo >&2 "Postgres is unavailable - sleeping"
  sleep 1
done

echo >&2 "Postgres is up - executing command"
exec $cmd
