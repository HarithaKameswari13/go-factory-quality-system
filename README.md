# Go Factory Quality System

A containerized, distributed factory quality monitoring backend built with **Go, gRPC, REST APIs, MySQL and Docker Compose**.

The system receives manufacturing measurements through an HTTP API, delegates quality validation to an independent gRPC service, and asynchronously persists the results using Go channels and background workers.

The project explores service decomposition, inter-service communication, concurrent processing, database integration and containerized deployment.

## Architecture

The application consists of three independently running services.

| Component | Responsibility | Technology |
|---|---|---|
| Receiver | Exposes REST endpoints, accepts measurements and coordinates processing | Go, Gorilla Mux |
| Checker | Validates measurements against predefined quality ranges | Go, gRPC |
| Database | Stores measurement results and vehicle history | MySQL |

### System Architecture

```text
                 External Client
                       |
                       | HTTP / JSON
                       v
              +-------------------+
              |  Receiver Service |
              |                   |
              |   REST API        |
              +-------------------+
                       |
                       | Submit measurement
                       v
              +-------------------+
              |    Go Channel     |
              |  In-memory Queue  |
              +-------------------+
                       |
                       v
              +-------------------+
              | Background Worker |
              +-------------------+
                       |
                +------+------+
                |             |
                v             v
          +-----------+  +-----------+
          |  Checker  |  |   MySQL   |
          |   gRPC    |  | Database  |
          +-----------+  +-----------+
```

## Measurement Processing Flow

1. A client submits a manufacturing measurement through the REST API.
2. The receiver validates the incoming request and places the measurement into a Go channel.
3. The receiver responds with HTTP `202 Accepted`, indicating that the measurement has been queued for processing.
4. A background worker retrieves the measurement from the channel.
5. The worker sends a gRPC request to the Checker service.
6. The Checker evaluates the measurement against predefined component-specific acceptance ranges.
7. The worker receives the validation result and persists the measurement, including its validity status, in MySQL.
8. Clients can retrieve previously processed measurements through the vehicle history endpoint.

## Concurrency and Asynchronous Processing

The receiver separates HTTP request handling from measurement validation and database persistence.

Instead of performing all operations within the HTTP request lifecycle, measurements are submitted to an in-memory Go channel and processed by background goroutines.

This design provides:

- Decoupling between request handling and background processing.
- Faster HTTP acknowledgements without waiting for database insertion.
- Independent processing of queued measurements.
- A foundation for concurrent processing through multiple workers.

The current implementation uses an in-memory queue. Consequently, queued measurements are not durable and can be lost if the receiver process terminates before processing them.

For production environments, a persistent messaging solution would be required.

## Inter-Service Communication

The Receiver and Checker are separate Go services communicating through gRPC.

The service contract is defined using Protocol Buffers in:

```text
proto/checker.proto
```

Generated Go protobuf and gRPC definitions are located in:

```text
internal/checkerpb/
```

The Checker exposes a measurement validation RPC. It receives the component name and measurement value and returns a validation result.

This separation keeps quality validation logic independent of HTTP request handling and allows the services to be developed and deployed separately.

## Quality Validation

The Checker maintains component-specific acceptance ranges.

Current example validation rules:

| Component | Minimum | Maximum |
|---|---:|---:|
| Wheel bolts | 20 | 25 |
| Battery | 20 | 25 |

These ranges are defined in the Checker service and can be modified independently of the Receiver.

## Database and Persistence

MySQL is used to maintain measurement history.

The database stores information such as:

- Vehicle identification number (VIN).
- Component name.
- Measured value.
- Validation result.
- Measurement processing timestamp.

Database initialization is handled through:

```text
db/init.sql
```

The Receiver uses Go's `database/sql` package and the MySQL driver to interact with the database.

Database connection details are supplied through environment variables rather than hardcoded into the application.

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Backend services and concurrent processing |
| Gorilla Mux | HTTP routing |
| gRPC | Inter-service communication |
| Protocol Buffers | Service contract and message definitions |
| MySQL | Persistent data storage |
| Go channels | In-memory measurement queue |
| Goroutines | Background processing |
| Docker | Containerization |
| Docker Compose | Multi-service orchestration |

## Project Structure

```text
go-factory-quality-system/
│
├── cmd/
│   ├── receiver/
│   │   └── main.go
│   │
│   └── checker/
│       └── main.go
│
├── internal/
│   └── checkerpb/
│       ├── checker.pb.go
│       └── checker_grpc.pb.go
│
├── proto/
│   └── checker.proto
│
├── db/
│   └── init.sql
│
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Getting Started

### Prerequisites

- Docker Desktop
- Docker Compose
- Git

### 1. Clone the repository

```bash
git clone https://github.com/HarithaKameswari13/go-factory-quality-system.git

cd go-factory-quality-system
```

### 2. Configure environment variables

Create a local `.env` file from the example:

**Windows PowerShell:**

```powershell
Copy-Item .env.example .env
```

**Linux / macOS:**

```bash
cp .env.example .env
```

Review the configuration before starting the services.

The example credentials are intended for local development only.

### 3. Start the services

```bash
docker compose up --build
```

Docker Compose builds and starts the Receiver, Checker and MySQL services.

The HTTP API is available at:

```text
http://localhost:8080
```

### 4. Stop the services

```bash
docker compose down
```

To also remove the database volume and its stored data:

```bash
docker compose down -v
```

**Warning:** Removing the volume permanently deletes the persisted database data.

## API Reference

### Health Check

**GET** `/health`

Checks the receiver and database availability.

Example:

```bash
curl http://localhost:8080/health
```

### Submit a Measurement

**POST** `/measurements`

Accepts a manufacturing measurement for asynchronous validation and storage.

Example request:

```json
{
  "vin": "VIN123",
  "component": "wheel_bolts",
  "value": 22
}
```

Example response:

```json
{
  "status": "accepted",
  "message": "Measurement queued for checking and storage"
}
```

HTTP status: `202 Accepted`

A successful response indicates that the measurement was accepted into the processing queue. It does not guarantee that validation and database persistence have completed.

### Retrieve Vehicle Measurement History

**GET** `/cars/{vin}/history`

Retrieves the stored measurement history for a specific vehicle.

Example:

```bash
curl http://localhost:8080/cars/VIN123/history
```

The response includes the recorded measurements and their validation results.

## Design Decisions and Engineering Trade-offs

### 1. Asynchronous processing

Background processing decouples HTTP request handling from validation and database operations.

This improves request responsiveness but introduces eventual processing and delivery considerations.

### 2. Service separation

The Receiver and Checker have separate responsibilities and communicate through a defined gRPC contract.

This reduces coupling between the API layer and business validation logic.

### 3. In-memory queue

Go channels provide a lightweight mechanism for coordinating concurrent workers without introducing an external messaging dependency.

The trade-off is that queued messages are not persisted across process restarts.

### 4. Environment-based configuration

Database credentials and connection settings are supplied through environment variables.

This keeps configuration separate from application source code and allows different environments to use different settings.



**Project:** Go Factory Quality System  
**Language:** Go  
**Repository:** [HarithaKameswari13/go-factory-quality-system](https://github.com/HarithaKameswari13/go-factory-quality-system)
