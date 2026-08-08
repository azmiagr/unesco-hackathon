FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk --no-cache add build-base pkgconf libwebp-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o unesco-hackathon ./cmd/app/main.go

FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata libwebp libstdc++

COPY --from=builder /app/unesco-hackathon .

EXPOSE 8081

CMD ["./unesco-hackathon"]
