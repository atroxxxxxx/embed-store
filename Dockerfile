# build configuration
FROM golang:1.25.6 AS builder
LABEL authors="atroxxx"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/service

# run configuration
FROM ubuntu:latest
LABEL authors="atroxxx"
WORKDIR /

COPY --from=builder /service /service
EXPOSE 8080

ENTRYPOINT ["/service"]
