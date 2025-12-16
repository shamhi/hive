package mapping

import (
	pbCommon "hive/gen/common"
	pb "hive/gen/store"
	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/domain/store"
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

func StoreToProto(s *store.Store) *pb.Store {
	if s == nil {
		return nil
	}
	return &pb.Store{
		StoreId:  s.ID,
		Name:     s.Name,
		Address:  s.Address,
		Location: LocationToProto(&s.Location),
	}
}
