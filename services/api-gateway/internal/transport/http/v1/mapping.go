package v1

import (
	"hive/services/api-gateway/internal/domain/assignment"
	"hive/services/api-gateway/internal/domain/base"
	"hive/services/api-gateway/internal/domain/drone"
	"hive/services/api-gateway/internal/domain/store"
)

func toBaseDTO(b *base.Base) BaseDTO {
	return BaseDTO{
		BaseID:  b.ID,
		Name:    b.Name,
		Address: b.Address,
		Location: Location{
			Lat: b.Location.Lat,
			Lon: b.Location.Lon,
		},
	}
}

func toStoreDTO(s *store.Store) StoreDTO {
	return StoreDTO{
		StoreID: s.ID,
		Name:    s.Name,
		Address: s.Address,
		Location: Location{
			Lat: s.Location.Lat,
			Lon: s.Location.Lon,
		},
	}
}

func toDroneDTO(d *drone.Drone, a *assignment.Assignment) DroneDTO {
	dDTO := DroneDTO{
		DroneID:             d.ID,
		Battery:             d.Battery,
		SpeedMps:            d.SpeedMps,
		ConsumptionPerMeter: d.ConsumptionPerMeter,
		Status:              string(d.Status),
		Location: Location{
			Lat: d.Location.Lat,
			Lon: d.Location.Lon,
		},
		UpdatedAtMs: d.UpdatedAt,
	}
	if a != nil {
		var tloc Location
		if a.Target != nil {
			tloc = Location{
				Lat: a.Target.Lat,
				Lon: a.Target.Lon,
			}
		}
		dDTO.Assignment = &AssignmentDTO{
			AssignmentID:   a.ID,
			Status:         string(a.Status),
			TargetLocation: &tloc,
		}
	}

	return dDTO
}
