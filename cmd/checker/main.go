package main

import (
	"context"
	"log"
	"net"
	"os"

	"factory-quality-system/internal/checkerpb"

	"google.golang.org/grpc"
)

type server struct {
	checkerpb.UnimplementedCheckerServer
}

func (s *server) CheckMeasurement(
	ctx context.Context,
	req *checkerpb.CheckRequest,
) (*checkerpb.CheckResponse, error) {

	valid := checkMeasurement(req.Component, req.Value)

	return &checkerpb.CheckResponse{
		Valid: valid,
	}, nil
}
func checkMeasurement(component string, value float64) bool {
	ranges := map[string][2]float64{
		"wheel_bolts": {20, 25},
		"battery":     {20, 25},
	}

	limits, known := ranges[component]

	return known && value >= limits[0] && value <= limits[1]
}

func resultMessage(known, valid bool) string {
	if !known {
		return "Unknown component"
	}
	if !valid {
		return "Measurement is outside the accepted range"
	}
	return "Measurement is within the accepted range"
}
func main() {
	grpcPort := os.Getenv("GRPC_PORT")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	checkerpb.RegisterCheckerServer(grpcServer, &server{})

	log.Println("Checker gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
