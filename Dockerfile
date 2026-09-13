# Build stage: full Go toolchain, static binary (no cgo at runtime).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

# Run stage: minimal image with just the binary.
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
