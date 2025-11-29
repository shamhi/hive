package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "hive/gen/telemetry"

	"google.golang.org/grpc"
)

// ConnectionManager хранит активные стримы для каждого drone_id
type ConnectionManager struct {
	// Ключ: drone_id, Значение: stream (send-only часть стрима)
	streams sync.Map
}

// Add регистрирует новое подключение дрона
func (cm *ConnectionManager) Add(droneID string, stream pb.TelemetryService_LinkServer) {
	cm.streams.Store(droneID, stream)
	log.Printf("🔌 Drone [%s] connected. Stream stored.", droneID)
}

// Remove удаляет подключение при разрыве
func (cm *ConnectionManager) Remove(droneID string) {
	cm.streams.Delete(droneID)
	log.Printf("❌ Drone [%s] disconnected. Stream removed.", droneID)
}

func (cm *ConnectionManager) Send(droneID string, cmd *pb.ServerCommand) error {
	val, ok := cm.streams.Load(droneID)
	if !ok {
		return fmt.Errorf("drone %s not connected", droneID)
	}

	stream, ok := val.(pb.TelemetryService_LinkServer)
	if !ok {
		return errors.New("invalid stream type")
	}

	// Thread-safe отправка в стрим (gRPC stream.Send не всегда потокобезопасен, но здесь допустим упрощение)
	if err := stream.Send(cmd); err != nil {
		return fmt.Errorf("failed to send to stream: %v", err)
	}

	log.Printf("📤 Sent command [%s] to drone [%s]", cmd.Action, droneID)
	return nil
}

// Server - реализация gRPC сервиса
type Server struct {
	pb.UnimplementedTelemetryServiceServer
	connManager *ConnectionManager
}

// 1. Обработка входящего стрима (Telemetry Flow)
func (s *Server) Link(stream pb.TelemetryService_LinkServer) error {
	var currentDroneID string

	// При выходе из функции (разрыв соединения) удаляем стрим из карты
	defer func() {
		if currentDroneID != "" {
			s.connManager.Remove(currentDroneID)
		}
	}()

	for {
		// Читаем сообщение от дрона
		msg, err := stream.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("Stream error: %v", err)
			return err
		}

		// --- ВАЛИДАЦИЯ ---
		if msg.Battery < 0 || msg.Battery > 100 {
			log.Printf("⚠️ Validation failed for %s: Battery %d%% out of range", msg.DroneId, msg.Battery)
			return errors.New("invalid battery level")
		}
		if msg.Lat < -90 || msg.Lat > 90 {
			return errors.New("invalid latitude")
		}

		// --- РЕГИСТРАЦИЯ СТРИМА ---
		// Сохраняем стрим при первом сообщении (когда узнали ID)
		if currentDroneID == "" {
			currentDroneID = msg.DroneId
			s.connManager.Add(currentDroneID, stream)
		}

		// --- ОБРАБОТКА ДАННЫХ (Эмуляция Kafka Producer) ---
		// Вместо producer.WriteMessage(...)
		log.Printf("📥 [TELEMETRY] %s | Bat: %d%% | Status: %s | Evt: %s",
			msg.DroneId, msg.Battery, msg.Status, msg.Event)

		// Если пришло спец-событие (ARRIVED, DELIVERED), логируем это отдельно
		// Dispatch Service должен будет прочитать это из Kafka
		if msg.Event != pb.DroneEvent_NONE {
			log.Printf("🔔 EVENT DETECTED: %s sent event %s -> producing to Kafka...", msg.DroneId, msg.Event)
		}

		// --- ACK (Опционально) ---
		// Мы не отправляем ACK явно на каждый пакет, чтобы не забивать канал,
		// но gRPC сам хэндлит flow control.
	}
}

// 2. Отправка команды дрону (вызывается из Dispatch Service)
func (s *Server) SendCommand(ctx context.Context, req *pb.DispatchCommandRequest) (*pb.DispatchCommandResponse, error) {
	log.Printf("📞 RPC SendCommand received: Send %s to %s", req.Action, req.DroneId)

	// Формируем команду для дрона
	cmd := &pb.ServerCommand{
		CommandId: "cmd-" + req.DroneId, // Генерация ID
		Action:    req.Action,
		Target:    req.Target,
	}

	// Пытаемся отправить через сохраненный стрим
	if err := s.connManager.Send(req.DroneId, cmd); err != nil {
		log.Printf("❌ Failed to send command: %v", err)
		return &pb.DispatchCommandResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.DispatchCommandResponse{Success: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Инициализация сервиса
	srv := &Server{
		connManager: &ConnectionManager{},
	}

	pb.RegisterTelemetryServiceServer(s, srv)

	log.Println("🚀 Telemetry Service running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
