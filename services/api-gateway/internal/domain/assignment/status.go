package assignment

type AssignmentStatus string

const (
	AssignmentStatusCreated        AssignmentStatus = "CREATED"
	AssignmentStatusAssigned       AssignmentStatus = "ASSIGNED"
	AssignmentStatusFlyingToStore  AssignmentStatus = "FLYING_TO_STORE"
	AssignmentStatusAtStore        AssignmentStatus = "AT_STORE"
	AssignmentStatusPickedUpCargo  AssignmentStatus = "PICKED_UP_CARGO"
	AssignmentStatusFlyingToClient AssignmentStatus = "FLYING_TO_CLIENT"
	AssignmentStatusAtClient       AssignmentStatus = "AT_CLIENT"
	AssignmentStatusDroppedCargo   AssignmentStatus = "DROPPED_CARGO"
	AssignmentStatusReturningBase  AssignmentStatus = "RETURNING_BASE"
	AssignmentStatusCompleted      AssignmentStatus = "COMPLETED"
	AssignmentStatusFailed         AssignmentStatus = "FAILED"
)
