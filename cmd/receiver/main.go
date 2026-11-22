package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"factory-quality-system/internal/checkerpb"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type MeasurementResult struct {
	VIN       string
	Component string
	Value     float64
	Valid     bool
}

func main() {
	_ = godotenv.Load()

	mysqlDSN := os.Getenv("MYSQL_DSN")

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Println("Connected to MySQL database!")
	queueSize, _ := strconv.Atoi(os.Getenv("QUEUE_SIZE"))
	measurementChannel := make(chan MeasurementResult, queueSize)
	workerCount, _ := strconv.Atoi(os.Getenv("WORKERS"))

	for i := 0; i < workerCount; i++ {
		go worker(measurementChannel, db)
	}

	conn, err := grpc.Dial(os.Getenv("GRPC_ADDRESS"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to checker: %v", err)
	}
	defer conn.Close()
	checkerClient := checkerpb.NewCheckerClient(conn)
	router := mux.NewRouter()

	router.HandleFunc("/measurements", func(w http.ResponseWriter, r *http.Request) {
		receiveMeasurement(w, r, checkerClient, measurementChannel)
	}).Methods("POST")

	log.Println("Receiver HTTP server listening on :8080")

	httpPort := os.Getenv("HTTP_PORT")
	log.Fatal(http.ListenAndServe(":"+httpPort, router))
}
func worker(measurementChannel <-chan MeasurementResult, db *sql.DB) {
	for measurement := range measurementChannel {
		log.Println("Saving measurement:", measurement.VIN)

		_, err := db.Exec(
			`INSERT INTO measurements (vin, component, measured_value, is_valid)
			VALUES (?, ?, ?, ?)`,
			measurement.VIN,
			measurement.Component,
			measurement.Value,
			measurement.Valid,
		)

		if err != nil {
			log.Println("Failed to save measurement:", err)
		}
	}
}

func receiveMeasurement(
	w http.ResponseWriter,
	r *http.Request,
	checkerClient checkerpb.CheckerClient,
	measurementChannel chan<- MeasurementResult,
) {
	var measurement struct {
		VIN       string  `json:"vin"`
		Component string  `json:"component"`
		Value     float64 `json:"value"`
	}

	err := json.NewDecoder(r.Body).Decode(&measurement)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received measurement: VIN=%s, Component=%s, Value=%f",
		measurement.VIN,
		measurement.Component,
		measurement.Value,
	)
	response, err := checkerClient.CheckMeasurement(
		r.Context(),
		&checkerpb.CheckRequest{
			Component: measurement.Component,
			Value:     measurement.Value,
		},
	)

	if err != nil {
		log.Printf("gRPC error: %v", err)
		http.Error(w, "Checker service failed", http.StatusInternalServerError)
		return
	}
	measurementChannel <- MeasurementResult{
		VIN:       measurement.VIN,
		Component: measurement.Component,
		Value:     measurement.Value,
		Valid:     response.Valid,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
