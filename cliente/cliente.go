package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	emergenciaspb "github.com/Ch3k00/tarea-2-sd/proto/emergenciaspb"

	"google.golang.org/grpc"
)

// Struct local para leer JSON
type EmergenciaInput struct {
	Name      string `json:"name"`
	Latitude  int32  `json:"latitude"`
	Longitude int32  `json:"longitude"`
	Magnitude int32  `json:"magnitude"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Uso: go run cliente.go <archivo.json>")
	}

	archivo := os.Args[1]
	file, err := os.ReadFile(archivo)
	if err != nil {
		log.Fatalf("Error leyendo archivo JSON: %v", err)
	}

	var emergenciasInput []EmergenciaInput
	if err := json.Unmarshal(file, &emergenciasInput); err != nil {
		log.Fatalf("Error al parsear JSON: %v", err)
	}

	// Conexión gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("No se pudo conectar con el servidor: %v", err)
	}
	defer conn.Close()

	client := emergenciaspb.NewServicioAsignacionClient(conn)

	// Convertir a mensaje gRPC
	var emergencias []*emergenciaspb.Emergencia
	for _, e := range emergenciasInput {
		emergencias = append(emergencias, &emergenciaspb.Emergencia{
			Name:      e.Name,
			Latitude:  e.Latitude,
			Longitude: e.Longitude,
			Magnitude: e.Magnitude,
		})
	}

	req := &emergenciaspb.EmergenciasRequest{Emergencias: emergencias}

	_, err = client.EnviarEmergencias(context.Background(), req)
	if err != nil {
		log.Fatalf("Error al enviar emergencias: %v", err)
	}

	log.Println("✅ Emergencias enviadas exitosamente al servidor.")
}
