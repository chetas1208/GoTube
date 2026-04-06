#!/bin/bash
set -e

DIRECTION=${1:-up}
DB_URL="${DATABASE_URL:-postgres://${POSTGRES_USER:-gotube}:${POSTGRES_PASSWORD:-gotube_secret}@${POSTGRES_HOST:-localhost}:${POSTGRES_PORT:-5432}/${POSTGRES_DB:-gotube_lite}?sslmode=disable}"
MIGRATIONS_PATH="./backend/api/migrations"

echo "Running migrations ${DIRECTION}..."
echo "Database: ${DB_URL}"

if command -v migrate &> /dev/null; then
  migrate -path "${MIGRATIONS_PATH}" -database "${DB_URL}" -verbose "${DIRECTION}"
else
  echo "golang-migrate not found. Install it:"
  echo "  brew install golang-migrate"
  echo "  or: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
  exit 1
fi

echo "Migrations ${DIRECTION} complete."
