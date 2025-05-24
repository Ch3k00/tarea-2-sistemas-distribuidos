package main

import (
	"context"
	"log"
	"math"
	"net"

	emergenciaspb "tarea-2-sd/proto/emergenciaspb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type servidorAsignacion struct {
	emergenciaspb.UnimplementedServicioAsignacionServer
	mongoClient  *mongo.Client
	dronesClient emergenciaspb.ServicioDronesClient
}

func (s *servidorAsignacion) mustEmbedUnimplementedServicioAsignacionServer() {}

func distancia(x1, y1, x2, y2 int32) float64 {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return math.Sqrt(dx*dx + dy*dy)
}

func (s *servidorAsignacion) obtenerDronMasCercano(ctx context.Context, lat, lon int32) (map[string]interface{}, error) {
	dronesColl := s.mongoClient.Database("emergencias").Collection("drones")
	cursor, err := dronesColl.Find(ctx, bson.M{"status": "available"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dronCercano map[string]interface{}
	distMin := math.MaxFloat64

	for cursor.Next(ctx) {
		var dron map[string]interface{}
		if err := cursor.Decode(&dron); err != nil {
			continue
		}
		dLat := int32(dron["latitude"].(float64))
		dLon := int32(dron["longitude"].(float64))
		dist := distancia(lat, lon, dLat, dLon)
		if dist < distMin {
			distMin = dist
			dronCercano = dron
		}
	}
	return dronCercano, nil
}

func (s *servidorAsignacion) EnviarEmergencias(ctx context.Context, req *emergenciaspb.EmergenciasRequest) (*emptypb.Empty, error) {
	for _, emergencia := range req.Emergencias {
		log.Printf("📥 Recibida emergencia: %s en (%d,%d), magnitud %d",
			emergencia.Name, emergencia.Latitude, emergencia.Longitude, emergencia.Magnitude)

		dron, err := s.obtenerDronMasCercano(ctx, emergencia.Latitude, emergencia.Longitude)
		if err != nil || dron == nil {
			log.Printf("❌ No hay drones disponibles para %s", emergencia.Name)
			continue
		}

		dronID, ok := dron["id"].(string)
		if !ok {
			log.Printf("❌ Dron sin ID válido: %v", dron)
			continue
		}

		log.Printf("✅ Dron asignado: %s → %s", dronID, emergencia.Name)

		// Crear nueva emergencia con dron asignado
		emergenciaConDron := &emergenciaspb.Emergencia{
			Name:      emergencia.Name,
			Latitude:  emergencia.Latitude,
			Longitude: emergencia.Longitude,
			Magnitude: emergencia.Magnitude,
			DronId:    dronID,
		}

		// Enviar al servicio de drones
		_, err = s.dronesClient.AsignarEmergencia(ctx, emergenciaConDron)
		if err != nil {
			log.Printf("❌ Error al enviar a drones: %v", err)
		} else {
			log.Printf("📡 Emergencia enviada al servicio de drones correctamente.")
		}
	}
	return &emptypb.Empty{}, nil
}

func main() {
	ctx := context.Background()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("❌ Error conectando a MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	connDrones, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ No se pudo conectar al servicio de drones: %v", err)
	}
	defer connDrones.Close()

	dronesClient := emergenciaspb.NewServicioDronesClient(connDrones)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ No se pudo escuchar en el puerto 50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	asignador := &servidorAsignacion{
		mongoClient:  mongoClient,
		dronesClient: dronesClient,
	}
	emergenciaspb.RegisterServicioAsignacionServer(grpcServer, asignador)

	log.Println("🚀 Servidor de asignación escuchando en puerto 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Error en gRPC: %v", err)
	}
}
