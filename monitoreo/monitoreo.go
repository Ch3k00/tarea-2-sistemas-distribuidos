package main

import (
	"encoding/json"
	"log"
	"net"

	emergenciaspb "tarea-2-sd/proto/emergenciaspb"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type servidorMonitoreo struct {
	emergenciaspb.UnimplementedServicioMonitoreoServer
	suscriptores []emergenciaspb.ServicioMonitoreo_RecibirActualizacionesServer
}

func (s *servidorMonitoreo) agregarSuscriptor(stream emergenciaspb.ServicioMonitoreo_RecibirActualizacionesServer) {
	s.suscriptores = append(s.suscriptores, stream)
	log.Println("✅ Cliente suscrito al monitoreo.")
}

func (s *servidorMonitoreo) emitirActualizacion(mensaje string) {
	for _, sub := range s.suscriptores {
		err := sub.Send(&emergenciaspb.Actualizacion{Mensaje: mensaje})
		if err != nil {
			log.Printf("❌ Error enviando actualización: %v", err)
		}
	}
}

// gRPC: método que deja a un cliente suscribirse al stream de actualizaciones
func (s *servidorMonitoreo) RecibirActualizaciones(_ *emptypb.Empty, stream emergenciaspb.ServicioMonitoreo_RecibirActualizacionesServer) error {
	s.agregarSuscriptor(stream)
	select {} // Mantiene la conexión abierta mientras el cliente escuche
}

func main() {
	// Conexión a RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ Error al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Error al abrir canal RabbitMQ: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare("monitoreo_q", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Error declarando cola: %v", err)
	}

	msgs, err := ch.Consume("monitoreo_q", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Error al consumir cola: %v", err)
	}

	// Crear instancia del servicio
	monitoreoSrv := &servidorMonitoreo{}

	// gRPC server
	grpcServer := grpc.NewServer()
	emergenciaspb.RegisterServicioMonitoreoServer(grpcServer, monitoreoSrv)

	// Procesar mensajes desde RabbitMQ y emitirlos a los clientes
	go func() {
		for d := range msgs {
			var m map[string]string
			err := json.Unmarshal(d.Body, &m)
			if err != nil {
				log.Printf("❌ Error al decodificar JSON: %v", err)
				continue
			}
			monitoreoSrv.emitirActualizacion(m["mensaje"])
		}
	}()

	// Escuchar gRPC en puerto 50053
	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("❌ No se pudo escuchar en el puerto 50053: %v", err)
	}

	log.Println("📡 Servicio de monitoreo escuchando en puerto 50053...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Fallo en gRPC: %v", err)
	}
}
