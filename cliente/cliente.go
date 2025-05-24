package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	emergenciaspb "tarea-2-sd/proto/emergenciaspb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EmergenciaInput struct {
	Name      string `json:"name"`
	Latitude  int32  `json:"latitude"`
	Longitude int32  `json:"longitude"`
	Magnitude int32  `json:"magnitude"`
}

func enviarEmergencias(jsonPath string) {
	file, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("❌ Error leyendo archivo JSON: %v", err)
	}

	var emergenciasInput []EmergenciaInput
	if err := json.Unmarshal(file, &emergenciasInput); err != nil {
		log.Fatalf("❌ Error al parsear JSON: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("❌ No se pudo conectar al servicio de asignación: %v", err)
	}
	defer conn.Close()

	client := emergenciaspb.NewServicioAsignacionClient(conn)

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
		log.Fatalf("❌ Error al enviar emergencias: %v", err)
	}

	log.Println("✅ Emergencias enviadas exitosamente.")
}

func escucharActualizaciones() {
	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ No se pudo conectar al servicio de monitoreo: %v", err)
	}
	defer conn.Close()

	client := emergenciaspb.NewServicioMonitoreoClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.RecibirActualizaciones(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("❌ Error al iniciar stream de monitoreo: %v", err)
	}

	log.Println("📡 Escuchando actualizaciones en tiempo real...")

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("📴 Fin del stream de monitoreo.")
			break
		}
		if err != nil {
			log.Fatalf("❌ Error recibiendo actualización: %v", err)
		}

		log.Printf("📢 %s", msg.Mensaje)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Uso: go run cliente.go <archivo.json>")
	}

	go escucharActualizaciones() // en paralelo
	time.Sleep(1 * time.Second)  // dar tiempo para que se conecte
	enviarEmergencias(os.Args[1])

	// Mantener proceso vivo mientras llegan actualizaciones
	select {}
}
