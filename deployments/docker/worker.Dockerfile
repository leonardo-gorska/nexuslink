# deployments/docker/worker.Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/worker ./cmd/worker

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/bin/worker /bin/worker

# GeoLite2 BD can be bounded via volume or added via secure download in prod.
USER nonroot:nonroot
ENTRYPOINT ["/bin/worker"]
