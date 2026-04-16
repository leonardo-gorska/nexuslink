# deployments/docker/api.Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/bin/api /bin/api
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/bin/api"]
