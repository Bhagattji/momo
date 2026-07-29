# Multi-stage Dockerfile for momo CLI

# Build stage
FROM golang:1.24-bullseye AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/momo ./cmd

# Final minimal image
FROM gcr.io/distroless/cc-debian11
COPY --from=build /out/momo /usr/local/bin/momo
ENTRYPOINT ["/usr/local/bin/momo"]
