FROM golang:1.23-alpine

ENV GOCACHE=/go/cache
RUN mkdir -p /go/cache && chown 1000:1000 /go/cache

WORKDIR /backend
RUN chown -R 1000:1000 /backend
USER 1000
# Install development tools

COPY go.mod go.sum  ./
RUN go mod download
RUN go install github.com/air-verse/air@latest

COPY . .
