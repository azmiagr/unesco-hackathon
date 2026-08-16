#!/bin/bash
set -e

APP_DIR="${APP_DIR:-$HOME/project/un/unesco-hackathon}"

echo "Changing directory to: $APP_DIR"
cd "$APP_DIR"

if [ -f .env ]; then
  set -a
  . ./.env
  set +a
fi

IMAGE_TAG="${IMAGE_TAG:-latest}"
export IMAGE_TAG
IMAGE="ghcr.io/${GITHUB_REPOSITORY:-azmiagr/unesco-hackathon}:${IMAGE_TAG}"
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

echo "Waiting for application health..."
for attempt in $(seq 1 30); do
  app_container="$($COMPOSE ps -q app)"
  if [ -n "$app_container" ]; then
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$app_container")"
    if [ "$status" = "healthy" ]; then
      break
    fi
    if [ "$status" = "exited" ] || [ "$status" = "dead" ]; then
      echo "Application container stopped unexpectedly."
      $COMPOSE logs --tail=100 app
      exit 1
    fi
  fi

  if [ "$attempt" = "30" ]; then
    echo "Application did not become healthy in time."
    $COMPOSE logs --tail=100 app
    exit 1
  fi
  sleep 2
done

echo "Cleaning up unused images..."
docker image prune -f

echo "Deploy complete."
$COMPOSE ps

echo "App: http://localhost:${APP_PORT}"
echo "Database: 127.0.0.1:3307"
