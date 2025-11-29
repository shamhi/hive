package main

import (
	"context"
	"log"
	"sync"
	"time"

	pb "hive/gen/telemetry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ServerAddr = "localhost:50051"
)

func runDrone(id string, lat, lon float64) {
	conn, err := grpc.Dial(ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[%s] connection failed: %v", id, err)
		return
	}
	defer conn.Close()

	client := pb.NewTelemetryServiceClient(conn)
	stream, err := client.Link(context.Background())
	if err != nil {
		log.Printf("[%s] stream creation failed: %v", id, err)
		return
	}
	log.Printf("[%s] connected", id)

	// Канал для синхронизации завершения работы горутин
	done := make(chan struct{})

	// Чтение команд от сервера
	go func() {
		defer close(done)
		for {
			cmd, err := stream.Recv()
			if err != nil {
				log.Printf("[%s] disconnected: %v", id, err)
				return
			}
			log.Printf("[%s] received command: %v", id, cmd.Action)

			// Место для вашей логики обработки
		}
	}()

	// Отправка телеметрии
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		battery := int32(100)

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				req := &pb.DroneTelemetry{
					DroneId:   id,
					Lat:       lat,
					Lon:       lon,
					Battery:   battery,
					Status:    pb.DroneStatus_FREE,
					Timestamp: time.Now().Unix(),
				}

				if err := stream.Send(req); err != nil {
					log.Printf("[%s] send error: %v", id, err)
					return
				}
			}
		}
	}()

	<-done
}

func main() {
	var wg sync.WaitGroup

	drones := []struct {
		id  string
		lat float64
		lon float64
	}{
		{"drone-1", 55.7500, 37.6100},
		{"drone-2", 55.7510, 37.6110},
		{"drone-3", 55.7520, 37.6120},
	}

	for _, d := range drones {
		wg.Add(1)
		go func(id string, lat, lon float64) {
			defer wg.Done()
			runDrone(id, lat, lon)
		}(d.id, d.lat, d.lon)
	}

	wg.Wait()

	log.Printf("All %d drones completed work", len(drones))
}
