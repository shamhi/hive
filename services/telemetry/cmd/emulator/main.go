package main

import (
	"context"
	"flag"
	"fmt"
	pbCommon "hive/gen/common"
	pb "hive/gen/telemetry"
	"hive/pkg/geo"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// диапазон начального заряда
	initialBatteryMin = 80.0  // проценты
	initialBatteryMax = 100.0 // проценты

	// диапазон скорости
	speedMinMps = 5.0  // м/с
	speedMaxMps = 15.0 // м/с

	// диапазон расхода батареи (процентов на метр)
	consumptionMinPerMeter = 0.001
	consumptionMaxPerMeter = 0.005

	// скорость зарядки (процентов в секунду)
	chargeRatePerSecond = 0.25

	// дистанция, погрешность для достижения цели
	arrivalThresholdMeters = 1.0

	// параметры по умолчанию
	defaultTelemetryAddress = "telemetry:50050"
	defaultPeriod           = 500 * time.Millisecond
	defaultDronesCount      = 10
)

type droneState struct {
	mu sync.Mutex

	id string

	lat float64
	lon float64

	targetLat float64
	targetLon float64
	hasTarget bool

	battery         float64
	speedMps        float64
	consumptionPerM float64
	status          pb.DroneStatus
	lastEvent       pb.DroneEvent
}

func newDroneState() *droneState {
	id := uuid.NewString()
	lat, lon := geo.RandMoscowPoint()
	battery := initialBatteryMin + rand.Float64()*(initialBatteryMax-initialBatteryMin)
	speedMps := speedMinMps + rand.Float64()*(speedMaxMps-speedMinMps)
	consumptionPerM := consumptionMinPerMeter + rand.Float64()*(consumptionMaxPerMeter-consumptionMinPerMeter)

	return &droneState{
		id:              id,
		lat:             lat,
		lon:             lon,
		battery:         battery,
		speedMps:        speedMps,
		consumptionPerM: consumptionPerM,
		status:          pb.DroneStatus_STATUS_FREE,
		lastEvent:       pb.DroneEvent_EVENT_NONE,
	}
}

func (d *droneState) applyCommand(cmd *pb.ServerCommand) {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch cmd.GetAction() {
	case pb.DroneAction_ACTION_WAIT:
		d.status = pb.DroneStatus_STATUS_FREE
	case pb.DroneAction_ACTION_FLY_TO:
		if cmd.GetTarget() != nil {
			d.targetLat = cmd.GetTarget().GetLat()
			d.targetLon = cmd.GetTarget().GetLon()
			d.hasTarget = true
			d.status = pb.DroneStatus_STATUS_BUSY
		}
	case pb.DroneAction_ACTION_PICKUP_CARGO:
		d.lastEvent = pb.DroneEvent_EVENT_PICKED_UP_CARGO
	case pb.DroneAction_ACTION_DROP_CARGO:
		d.lastEvent = pb.DroneEvent_EVENT_DROPPED_CARGO
	case pb.DroneAction_ACTION_CHARGE:
		d.status = pb.DroneStatus_STATUS_CHARGING
	default:
	}
}

func (d *droneState) step(dt time.Duration) *pb.DroneTelemetry {
	d.mu.Lock()
	defer d.mu.Unlock()

	seconds := dt.Seconds()

	if d.hasTarget && d.status == pb.DroneStatus_STATUS_BUSY {
		dist := geo.HaversineDistance(d.lat, d.lon, d.targetLat, d.targetLon)
		if dist < arrivalThresholdMeters {
			d.hasTarget = false
			d.lastEvent = guessArrivalEvent(d.lastEvent)
		} else {
			move := math.Min(dist, d.speedMps*seconds)
			if dist > 0 {
				k := move / dist
				d.lat += (d.targetLat - d.lat) * k
				d.lon += (d.targetLon - d.lon) * k
			}
			d.battery -= move * d.consumptionPerM
			if d.battery < 0 {
				d.battery = 0
			}
		}
	} else if d.status == pb.DroneStatus_STATUS_CHARGING {
		d.battery += chargeRatePerSecond * seconds
		if d.battery >= 100 {
			d.battery = 100
			d.status = pb.DroneStatus_STATUS_FREE
			d.lastEvent = pb.DroneEvent_EVENT_FULLY_CHARGED
		}
	}

	ev := d.lastEvent
	d.lastEvent = pb.DroneEvent_EVENT_NONE

	return &pb.DroneTelemetry{
		DroneId: d.id,
		DroneLocation: &pbCommon.Location{
			Lat: d.lat,
			Lon: d.lon,
		},
		Battery:   d.battery,
		Status:    d.status,
		Timestamp: time.Now().UnixMilli(),
		Event:     ev,
	}
}

func guessArrivalEvent(prev pb.DroneEvent) pb.DroneEvent {
	switch prev {
	case pb.DroneEvent_EVENT_NONE:
		return pb.DroneEvent_EVENT_ARRIVED_AT_STORE
	case pb.DroneEvent_EVENT_PICKED_UP_CARGO:
		return pb.DroneEvent_EVENT_ARRIVED_AT_CLIENT
	case pb.DroneEvent_EVENT_DROPPED_CARGO:
		return pb.DroneEvent_EVENT_ARRIVED_AT_BASE
	default:
		return pb.DroneEvent_EVENT_NONE
	}
}

func runEmulator(ctx context.Context, addr string, period time.Duration) error {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	defer conn.Close()

	client := pb.NewTelemetryServiceClient(conn)

	stream, err := client.Link(ctx)
	if err != nil {
		return fmt.Errorf("failed to open gRPC stream: %w", err)
	}

	state := newDroneState()
	fmt.Printf("drone %s started at (%.6f, %.6f) with battery %.2f%%\n", state.id, state.lat, state.lon, state.battery)

	errCh := make(chan error, 2)

	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				errCh <- fmt.Errorf("drone %s receive error: %w", state.id, err)
				return
			}
			fmt.Printf("drone %s command: action=%s type=%s target=(%.6f, %.6f)\n",
				state.id,
				cmd.GetAction().String(),
				cmd.GetType().String(),
				cmd.GetTarget().GetLat(),
				cmd.GetTarget().GetLon(),
			)
			state.applyCommand(cmd)
		}
	}()

	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		last := time.Now()
		for {
			select {
			case <-ctx.Done():
				errCh <- nil
				return
			case now := <-ticker.C:
				dt := now.Sub(last)
				last = now
				tm := state.step(dt)
				if err := stream.Send(tm); err != nil {
					errCh <- fmt.Errorf("drone %s send error: %w", state.id, err)
					return
				}
			}
		}
	}()

	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

func main() {
	addr := flag.String("addr", defaultTelemetryAddress, "telemetry gRPC server address")
	period := flag.Duration("period", defaultPeriod, "telemetry send period")
	n := flag.Int("n", defaultDronesCount, "number of drones to emulate")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(*n)

	for range *n {
		go func() {
			defer wg.Done()

			if err := runEmulator(ctx, *addr, *period); err != nil && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "emulator error: %v\n", err)
				cancel()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-quitCtx.Done():
		fmt.Println("signal received, stopping...")
		cancel()
		<-done
	case <-done:
		fmt.Println("all emulators stopped")
	}
}
