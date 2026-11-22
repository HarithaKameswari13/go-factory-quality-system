# Go Factory Quality System

A containerized factory quality monitoring system built with Go, gRPC, REST APIs, and MySQL.

It receives manufacturing measurements, validates them against predefined quality ranges, and asynchronously stores the results in a database using goroutines and channels. Docker Compose orchestrates the services.
=======
# Factory Quality-Control Backend (Go)

A small learning project demonstrating:
- **Gorilla Mux**: HTTP routing in the receiver service.
- **gRPC**: communication between the receiver and checker services.
- **Goroutines + channels**: background workers process queued measurements.
- **MySQL**: persistent measurement history.
- **Docker Compose**: starts all services together.

## Start everything

Install Docker Desktop (with Docker Compose), open a terminal in this folder, and run:

```bash
docker compose up --build
```

The first run downloads images and Go dependencies. When startup finishes, the receiver is available at `http://localhost:8080`.

Stop with `Ctrl+C`. To stop and remove the containers:

```bash
docker compose down
```

To also delete the stored database data (this permanently clears the demo history):

```bash
docker compose down -v
```

## Try it

### 1. Check the receiver

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

### 2. Submit a measurement

```bash
curl -X POST http://localhost:8080/measurements \
  -H "Content-Type: application/json" \
  -d '{"vin":"VIN123","component":"wheel_bolts","value":22}'
```

Expected response (HTTP 202):

```json
{"status":"accepted","message":"Measurement queued for checking and storage"}
```

**Important:** `202 Accepted` means the receiver queued the work. A background worker then asks the checker to validate it and saves the result to MySQL. Give it a moment before requesting history.

Try an invalid measurement:

```bash
curl -X POST http://localhost:8080/measurements \
  -H "Content-Type: application/json" \
  -d '{"vin":"VIN123","component":"wheel_bolts","value":22}'
```

Try an unknown component:

```bash
curl -X POST http://localhost:8080/measurements \
  -H "Content-Type: application/json" \
  -d '{"vin":"VIN123","component":"paint","value":22}'
```

### 3. Retrieve a car's history

```bash
curl http://localhost:8080/cars/VIN123/history
```

You should see the submitted measurements, their values, whether each was valid, and the time they were checked.

## API summary

| Method | Endpoint | Purpose |
|---|---|---|
| GET | `/health` | Check receiver and database |
| POST | `/measurements` | Queue a measurement |
| GET | `/cars/{vin}/history` | Read a car's saved measurements |

Example request body:

```json
{
  "vin": "VIN123",
  "component": "wheel_bolts",
  "value": 22
}
```

## How the code is organised

```text
cmd/
  receiver/       HTTP API, worker queue, MySQL access
  checker/        gRPC server containing validation rules
internal/
  checkerpb/      Small gRPC client/server definitions
proto/
  checker.proto   The service contract
db/
  init.sql        Creates the measurements table
docker-compose.yml
Dockerfile
```

## Follow the flow in the code

1. `cmd/receiver/main.go` creates the Gorilla Mux router and starts worker goroutines.
2. `submitMeasurement` validates the incoming JSON and puts a job on a Go channel.
3. `worker` takes jobs from the channel, calls the checker using gRPC, and inserts the result into MySQL.
4. `cmd/checker/main.go` receives the gRPC call and applies the allowed ranges.
5. `carHistory` queries MySQL and returns the records as JSON.

## Where are the rules?

Open `cmd/checker/main.go` and find the `ranges` map:

```go
ranges := map[string][2]float64{
    "wheel_bolts": {20, 25},
    "battery":     {20, 25},
}
```

Change these values to experiment with different acceptance ranges.

## Learning notes and limitations

- This is a teaching/demo project, not production-ready factory software.
- The receiver returns `202 Accepted` before background processing finishes. Check the history after a short pause.
- The queue is in memory. If the receiver stops before queued jobs are processed, those queued jobs are lost. A production system would use a durable message broker or database-backed queue.
- The demo uses an insecure gRPC connection inside the local Docker network. Production deployments should use TLS and authentication.
- The gRPC contract is defined in `proto/checker.proto`, with generated Go protobuf and gRPC code in `internal/checkerpb/`.
- The checker and receiver are separate services and can be developed/deployed independently.
>>>>>>> 503f5fc (Initial commit: Factory Quality System)
