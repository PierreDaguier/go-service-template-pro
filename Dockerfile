# syntax=docker/dockerfile:1.7

FROM golang:1.26 AS builder
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /bin/service ./cmd/service

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /bin/service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
