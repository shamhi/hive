package drone

import "hive/services/dispatch/internal/domain/shared"

type Target struct {
	Location *shared.Location
	Type     TargetType
}

type TargetType string

const (
	TargetTypeNone   TargetType = "none"
	TargetTypePoint  TargetType = "point"
	TargetTypeStore  TargetType = "store"
	TargetTypeClient TargetType = "client"
	TargetTypeBase   TargetType = "base"
)
