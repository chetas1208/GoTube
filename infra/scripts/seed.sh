#!/bin/bash
set -e

DB_URL="${DATABASE_URL:-postgres://${POSTGRES_USER:-gotube}:${POSTGRES_PASSWORD:-gotube_secret}@${POSTGRES_HOST:-localhost}:${POSTGRES_PORT:-5432}/${POSTGRES_DB:-gotube_lite}?sslmode=disable}"

echo "Seeding database..."

psql "${DB_URL}" << 'SQL'
-- Seed a demo user (password: "password123" hashed with bcrypt)
INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
VALUES (
  'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
  'demo',
  'demo@gotube.dev',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  now(),
  now()
) ON CONFLICT (email) DO NOTHING;

SELECT 'Seed complete: demo user created (demo@gotube.dev / password123)' AS status;
SQL
