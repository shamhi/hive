package drone

import "hive/services/telemetry/internal/domain/shared"

type ServerCommand struct {
	CommandID string
	DroneID   string
	Action    DroneAction
	Target    *shared.Location
	Type      TargetType
}
