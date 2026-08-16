FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk --no-cache add build-base pkgconf libwebp-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o unesco-hackathon ./cmd/app/main.go

FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata libwebp libstdc++ \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=builder --chown=app:app /app/unesco-hackathon .

EXPOSE 8081

USER app

CMD ["./unesco-hackathon"]
