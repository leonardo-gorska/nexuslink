# Load Testing Results

Due to environment restrictions, a real load test environment couldn't be spun up. However, the `scripts/loadtest.sh` script is provided to run `vegeta` and gather these specs on the host machine.

## Realistic Expected Throughput (Local Docker Compose)

Based on similar architectures (Go 1.22 + Chi + Postgres + Redis):

### Hot Path (GET `/r/{hash}`)
- **Cache Hit Rate:** ~95%
- **RPS Target:** 5,000 - 8,000 requests per second
- **Latency (p99):** < 5ms
- **Bottleneck:** Redis Network I/O / Go GC

### Write Path (POST `/api/v1/links`)
- **RPS Target:** ~500 requests per second
- **Latency (p99):** < 15ms
- **Bottleneck:** PostgreSQL Disk I/O (fsync)

## Executing the Load Test

```bash
# Ensure services are up
make docker-up

# Run 500 RPS for 15s against the API
chmod +x scripts/loadtest.sh
./scripts/loadtest.sh http://localhost:8080/r/aB3xK9z 15s 500
```
