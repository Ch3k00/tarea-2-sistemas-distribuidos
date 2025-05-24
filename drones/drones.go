package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	emergenciaspb "tarea-2-sd/proto/emergenciaspb"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type servidorDrones struct {
	emergenciaspb.UnimplementedServicioDronesServer
	connRabbit  *amqp.Connection
	mongoClient *mongo.Client
}

func (s *servidorDrones) mustEmbedUnimplementedServicioDronesServer() {}

func distancia(x1, y1, x2, y2 int32) float64 {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return math.Sqrt(dx*dx + dy*dy)
}

func publicarMensaje(conn *amqp.Connection, queue string, payload map[string]interface{}) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("❌ Error abriendo canal RabbitMQ para %s: %v", queue, err)
		return
	}
	defer ch.Close()

	ch.QueueDeclare(queue, false, false, false, false, nil)

	body, _ := json.Marshal(payload)
	err = ch.Publish("", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Printf("❌ Error publicando en %s: %v", queue, err)
	}
}

func publicarInicioRegistro(conn *amqp.Connection, emergencia *emergenciaspb.Emergencia) {
	payload := map[string]interface{}{
		"name":      emergencia.Name,
		"latitude":  emergencia.Latitude,
		"longitude": emergencia.Longitude,
		"magnitude": emergencia.Magnitude,
	}
	publicarMensaje(conn, "registro_q", payload)
}

func publicarFinRegistro(conn *amqp.Connection, emergencia *emergenciaspb.Emergencia) {
	payload := map[string]interface{}{
		"name": emergencia.Name,
	}
	publicarMensaje(conn, "registro_fin_q", payload)
}

func publicarActualizacion(conn *amqp.Connection, mensaje string) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("❌ Error creando canal para monitoreo: %v", err)
		return
	}
	defer ch.Close()

	ch.QueueDeclare("monitoreo_q", false, false, false, false, nil)

	body, _ := json.Marshal(map[string]string{"mensaje": mensaje})
	err = ch.Publish("", "monitoreo_q", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Printf("❌ Error publicando monitoreo: %v", err)
	}
}

func (s *servidorDrones) AsignarEmergencia(ctx context.Context, emergencia *emergenciaspb.Emergencia) (*emptypb.Empty, error) {
	log.Printf("🛫 Emergencia recibida: %s en (%d,%d) magnitud %d",
		emergencia.Name, emergencia.Latitude, emergencia.Longitude, emergencia.Magnitude)

	// Publicar inicio
	publicarInicioRegistro(s.connRabbit, emergencia)

	// Simular desplazamiento
	dist := distancia(0, 0, emergencia.Latitude, emergencia.Longitude)
	tiempoDesplazamiento := time.Duration(dist * 0.5 * float64(time.Second))
	log.Printf("🛰️ Desplazándose hacia %s (%.2f unidades)...", emergencia.Name, dist)

	for t := time.Duration(0); t < tiempoDesplazamiento; t += 5 * time.Second {
		time.Sleep(5 * time.Second)
		publicarActualizacion(s.connRabbit, fmt.Sprintf("Dron en camino a %s...", emergencia.Name))
	}
	time.Sleep(tiempoDesplazamiento % (5 * time.Second))

	// Simular apagado
	tiempoApagado := time.Duration(emergencia.Magnitude) * 2 * time.Second
	log.Printf("🔥 Apagando incendio %s (%v)...", emergencia.Name, tiempoApagado)

	for t := time.Duration(0); t < tiempoApagado; t += 5 * time.Second {
		time.Sleep(5 * time.Second)
		publicarActualizacion(s.connRabbit, fmt.Sprintf("Apagando %s...", emergencia.Name))
	}
	time.Sleep(tiempoApagado % (5 * time.Second))

	log.Printf("✅ Emergencia %s extinguida.", emergencia.Name)
	publicarActualizacion(s.connRabbit, fmt.Sprintf("Emergencia %s extinguida.", emergencia.Name))

	// Publicar finalización
	publicarFinRegistro(s.connRabbit, emergencia)

	// Actualizar posición final del dron en MongoDB
	ctxMongo := context.Background()
	dronesColl := s.mongoClient.Database("emergencias").Collection("drones")

	filter := bson.M{"id": "dron01"} // ⚠️ Reemplaza con ID dinámico más adelante
	update := bson.M{
		"$set": bson.M{
			"latitude":  emergencia.Latitude,
			"longitude": emergencia.Longitude,
			"status":    "available",
		},
	}

	_, err := dronesColl.UpdateOne(ctxMongo, filter, update)
	if err != nil {
		log.Printf("❌ Error actualizando posición del dron: %v", err)
	} else {
		log.Printf("📍 Posición del dron actualizada en MongoDB.")
	}

	return &emptypb.Empty{}, nil
}

func main() {
	// RabbitMQ
	connRabbit, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ No se pudo conectar a RabbitMQ: %v", err)
	}
	defer connRabbit.Close()

	// MongoDB
	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("❌ No se pudo conectar a MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// gRPC
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ No se pudo escuchar en el puerto 50052: %v", err)
	}

	grpcServer := grpc.NewServer()
	emergenciaspb.RegisterServicioDronesServer(grpcServer, &servidorDrones{
		connRabbit:  connRabbit,
		mongoClient: mongoClient,
	})

	log.Println("🚁 Servicio de drones escuchando en puerto 50052...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Error al iniciar servicio de drones: %v", err)
	}
}
