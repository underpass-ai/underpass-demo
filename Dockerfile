# syntax=docker/dockerfile:1.7
#
# Underpass Demo — Interactive TUI for the USS Underpass spaceship demo.
#
# Build:  docker build -t underpass-demo .
# Run:    docker run -it underpass-demo

FROM docker.io/library/golang:1.25-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/tlctl ./cmd/tlctl

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/tlctl /app/tlctl

ENTRYPOINT ["/app/tlctl", "--embedded"]
