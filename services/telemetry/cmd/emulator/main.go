package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	pbCommon "hive/gen/common"
	pb "hive/gen/telemetry"
	"hive/pkg/geo"
	"io"
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
	// диапазон начального заряда (%)
	initialBatteryMin = 80.0
	initialBatteryMax = 100.0

	// диапазон скорости (м/с)
	speedMinMps = 320.0
	speedMaxMps = 325.0

	// диапазон расхода батареи (%/м)
	consumptionMinPerMeter = 0.001
	consumptionMaxPerMeter = 0.005

	// скорость зарядки (%/с)
	chargeRatePerSecond = 0.25

	// дистанция, погрешность для достижения цели
	arrivalThresholdMeters = 1.0

	// параметры по умолчанию
	defaultTelemetryAddress = "passthrough:///telemetry:50050"
	defaultPeriod           = 500 * time.Millisecond
	defaultDronesCount      = 10
)

type droneState struct {
	mu sync.Mutex

	id string

	lat float64
	lon float64

	targetLat  float64
	targetLon  float64
	hasTarget  bool
	targetType pb.TargetType

	battery         float64
	speedMps        float64
	consumptionPerM float64
	status          pb.DroneStatus

	pendingEvent pb.DroneEvent
}

func newDroneState(r *rand.Rand) *droneState {
	id := uuid.NewString()
	lat, lon := geo.RandMoscowPoint()
	battery := initialBatteryMin + r.Float64()*(initialBatteryMax-initialBatteryMin)
	speedMps := speedMinMps + r.Float64()*(speedMaxMps-speedMinMps)
	consumptionPerM := consumptionMinPerMeter + r.Float64()*(consumptionMaxPerMeter-consumptionMinPerMeter)

	return &droneState{
		id:              id,
		lat:             lat,
		lon:             lon,
		battery:         battery,
		speedMps:        speedMps,
		consumptionPerM: consumptionPerM,
		status:          pb.DroneStatus_STATUS_FREE,
		pendingEvent:    pb.DroneEvent_EVENT_NONE,
		targetType:      pb.TargetType_TARGET_NONE,
	}
}

func (d *droneState) applyCommand(cmd *pb.ServerCommand) {
	if cmd == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	switch cmd.GetAction() {
	case pb.DroneAction_ACTION_WAIT:
		d.status = pb.DroneStatus_STATUS_FREE
		d.hasTarget = false
		d.targetType = pb.TargetType_TARGET_NONE

	case pb.DroneAction_ACTION_FLY_TO:
		if cmd.GetTarget() != nil {
			d.targetLat = cmd.GetTarget().GetLat()
			d.targetLon = cmd.GetTarget().GetLon()
			d.hasTarget = true
			d.targetType = cmd.GetType()
			d.status = pb.DroneStatus_STATUS_BUSY
		}

	case pb.DroneAction_ACTION_PICKUP_CARGO:
		d.pendingEvent = pb.DroneEvent_EVENT_PICKED_UP_CARGO

	case pb.DroneAction_ACTION_DROP_CARGO:
		d.pendingEvent = pb.DroneEvent_EVENT_DROPPED_CARGO

	case pb.DroneAction_ACTION_CHARGE:
		d.status = pb.DroneStatus_STATUS_CHARGING
		d.hasTarget = false
		d.targetType = pb.TargetType_TARGET_NONE

	default:
		// ignore
	}
}

func (d *droneState) step(dt time.Duration) *pb.DroneTelemetry {
	d.mu.Lock()
	defer d.mu.Unlock()

	seconds := dt.Seconds()

	if d.hasTarget && d.status == pb.DroneStatus_STATUS_BUSY {
		dist := geo.HaversineDistance(d.lat, d.lon, d.targetLat, d.targetLon)

		if dist <= arrivalThresholdMeters {
			d.hasTarget = false
			switch d.targetType {
			case pb.TargetType_TARGET_STORE:
				d.pendingEvent = pb.DroneEvent_EVENT_ARRIVED_AT_STORE
			case pb.TargetType_TARGET_CLIENT:
				d.pendingEvent = pb.DroneEvent_EVENT_ARRIVED_AT_CLIENT
			case pb.TargetType_TARGET_BASE:
				d.pendingEvent = pb.DroneEvent_EVENT_ARRIVED_AT_BASE
			default:
				// point/none
			}
			d.targetType = pb.TargetType_TARGET_NONE
		} else {
			move := math.Min(dist, d.speedMps*seconds)
			if dist > 0 {
				k := move / dist
				d.lat = d.lat + (d.targetLat-d.lat)*k
				d.lon = d.lon + (d.targetLon-d.lon)*k
			}

			d.battery -= move * d.consumptionPerM
			if d.battery < 0 {
				d.battery = 0
				d.hasTarget = false
				d.targetType = pb.TargetType_TARGET_NONE
				d.status = pb.DroneStatus_STATUS_FREE
			}
		}
	}

	if d.status == pb.DroneStatus_STATUS_CHARGING {
		if d.battery < 100.0 {
			d.battery += chargeRatePerSecond * seconds
			if d.battery >= 100.0 {
				d.battery = 100.0
				d.pendingEvent = pb.DroneEvent_EVENT_FULLY_CHARGED
				d.status = pb.DroneStatus_STATUS_FREE
			}
		}
	}

	evt := d.pendingEvent
	d.pendingEvent = pb.DroneEvent_EVENT_NONE

	return &pb.DroneTelemetry{
		DroneId: d.id,
		DroneLocation: &pbCommon.Location{
			Lat: d.lat,
			Lon: d.lon,
		},
		Battery:             d.battery,
		SpeedMps:            d.speedMps,
		ConsumptionPerMeter: d.consumptionPerM,
		Status:              d.status,
		Timestamp:           time.Now().UnixMilli(),
		Event:               evt,
	}
}

func runDrone(ctx context.Context, client pb.TelemetryServiceClient, st *droneState, period time.Duration) error {
	stream, err := client.Link(ctx)
	if err != nil {
		return fmt.Errorf("link(): %w", err)
	}

	recvErr := make(chan error, 1)
	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				recvErr <- err
				return
			}
			st.applyCommand(cmd)
		}
	}()

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	if err := stream.Send(st.step(period)); err != nil {
		return fmt.Errorf("send initial telemetry: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			_ = stream.CloseSend()
			return ctx.Err()

		case err := <-recvErr:
			if err == nil || err == io.EOF {
				return nil
			}
			return fmt.Errorf("recv command: %w", err)

		case <-ticker.C:
			t := st.step(period)
			if err := stream.Send(t); err != nil {
				return fmt.Errorf("send telemetry: %w", err)
			}
		}
	}
}

func main() {
	var (
		addr   string
		period time.Duration
		count  int
		seed   int64
	)

	flag.StringVar(&addr, "addr", defaultTelemetryAddress, "telemetry gRPC address")
	flag.DurationVar(&period, "period", defaultPeriod, "telemetry send period")
	flag.IntVar(&count, "count", defaultDronesCount, "number of drones")
	flag.Int64Var(&seed, "seed", time.Now().UnixNano(), "random seed")
	flag.Parse()

	if count <= 0 {
		fmt.Println("count must be > 0")
		os.Exit(2)
	}
	if period <= 0 {
		fmt.Println("period must be > 0")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("failed to connect telemetry: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewTelemetryServiceClient(conn)

	rootRand := rand.New(rand.NewSource(seed))

	var wg sync.WaitGroup
	wg.Add(count)

	errCh := make(chan error, count)

	fmt.Printf("telemetry emulator started: addr=%s period=%s drones=%d seed=%d\n", addr, period, count, seed)

	for i := 0; i < count; i++ {
		dr := rand.New(rand.NewSource(rootRand.Int63()))
		st := newDroneState(dr)

		go func(s *droneState) {
			defer wg.Done()
			if err := runDrone(ctx, client, s, period); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- fmt.Errorf("drone %s: %w", s.id, err)
			}
		}(st)
	}

	select {
	case err := <-errCh:
		fmt.Printf("emulator error: %v\n", err)
		stop()
	case <-ctx.Done():
		// graceful stop
	}

	wg.Wait()
	fmt.Println("telemetry emulator stopped")
}
