# Estagio 1: Build
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api-server ./cmd/api

# Estagio 2: Run
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/api-server .
EXPOSE 8080
CMD ["./api-server"]
