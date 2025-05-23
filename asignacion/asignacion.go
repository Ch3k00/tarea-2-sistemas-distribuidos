package main

import (
	"context"
	"log"
	"net"

	emergenciaspb "github.com/Ch3k00/tarea-2-sd/proto/emergenciaspb"

	"google.golang.org/grpc"
)

type servidorAsignacion struct {
	emergenciaspb.UnimplementedServicioAsignacionServer
}

func (s *servidorAsignacion) EnviarEmergencias(ctx context.Context, req *emergenciaspb.EmergenciasRequest) (*emergenciaspb.Empty, error) {
	for _, emergencia := range req.Emergencias {
		log.Printf("Recibida emergencia: %s en (%d,%d), magnitud %d\n",
			emergencia.Name, emergencia.Latitude, emergencia.Longitude, emergencia.Magnitude)
	}
	return &emergenciaspb.Empty{}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("No se pudo escuchar en el puerto 50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	emergenciaspb.RegisterServicioAsignacionServer(grpcServer, &servidorAsignacion{})

	log.Println("Servidor de asignación escuchando en puerto 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Fallo al iniciar servidor gRPC: %v", err)
	}
}
