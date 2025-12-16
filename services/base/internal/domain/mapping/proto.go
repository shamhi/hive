package mapping

import (
	pb "hive/gen/base"
	pbCommon "hive/gen/common"
	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"
)

func LocationToProto(loc *shared.Location) *pbCommon.Location {
	if loc == nil {
		return nil
	}
	return &pbCommon.Location{
		Lat: loc.Lat,
		Lon: loc.Lon,
	}
}

func BaseToProto(s *base.Base) *pb.Base {
	if s == nil {
		return nil
	}
	return &pb.Base{
		BaseId:   s.ID,
		Name:     s.Name,
		Address:  s.Address,
		Location: LocationToProto(&s.Location),
	}
}
