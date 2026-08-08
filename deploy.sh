#!/bin/bash
set -e

APP_DIR="${APP_DIR:-$HOME/un}"

IMAGE="ghcr.io/${GITHUB_REPOSITORY:-azmiagr/unesco-hackathon}:latest"
APP_PORT="${PORT:-8081}"

if docker compose version >/dev/null 2>&1; then
  COMPOSE="docker compose"
else
  COMPOSE="docker-compose"
fi

echo "Syncing latest config from repository..."
git pull origin main

echo "Pulling latest image: $IMAGE"
docker pull "$IMAGE"

echo "Restarting services..."
$COMPOSE down --remove-orphans
$COMPOSE up -d

echo "Cleaning up unused images..."
docker image prune -f

echo "Deploy complete."
$COMPOSE ps

echo "App: http://localhost:${APP_PORT}"
echo "Database: 127.0.0.1:3307"
