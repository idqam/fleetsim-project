package constants

type SpawnRequestStatus string

const (
	SpawnRequestStatusPending    SpawnRequestStatus = "pending"
	SpawnRequestStatusValidating SpawnRequestStatus = "validating"
	SpawnRequestStatusQueued     SpawnRequestStatus = "queued"
	SpawnRequestStatusProcessing SpawnRequestStatus = "processing"
	SpawnRequestStatusCompleted  SpawnRequestStatus = "completed"
	SpawnRequestStatusFailed     SpawnRequestStatus = "failed"
	SpawnRequestStatusCancelled  SpawnRequestStatus = "cancelled"
)

type VehicleClass string

const (
	VehicleClassFleet      VehicleClass = "fleet"
	VehicleClassBackground VehicleClass = "background"
)

type VehicleStatus string

const (
	VehicleStatusRequested VehicleStatus = "requested"
	VehicleStatusQueued    VehicleStatus = "queued"
	VehicleStatusSpawning  VehicleStatus = "spawning"

	VehicleStatusMoving    VehicleStatus = "moving"
	VehicleStatusIdle      VehicleStatus = "idle"
	VehicleStatusStopped   VehicleStatus = "stopped"
	VehicleStatusRefueling VehicleStatus = "refueling"

	VehicleStatusCompleted VehicleStatus = "completed"
	VehicleStatusRemoved   VehicleStatus = "removed"
	VehicleStatusFailed    VehicleStatus = "failed"
)

type VehicleType string

const (
	VehicleTypeTruck VehicleType = "truck"
	VehicleTypeVan   VehicleType = "van"
	VehicleTypeCar   VehicleType = "car"
)

type WSMessageType string

const (
	WSMsgVehiclePosition WSMessageType = "vehicle_position"
	WSMsgTrafficUpdate   WSMessageType = "traffic_update"
	WSMsgSimulationStart WSMessageType = "simulation_start"
	WSMsgSimulationStop  WSMessageType = "simulation_stop"
	WSMsgSimulationError WSMessageType = "simulation_error"
	WSMsgRoutingDecision WSMessageType = "routing_decision"

	WSMsgSpawnRequest     WSMessageType = "spawn_request"
	WSMsgSpawnResponse    WSMessageType = "spawn_response"
	WSMsgVehicleRemove    WSMessageType = "vehicle_remove"
	WSMsgUserVehiclesList WSMessageType = "user_vehicles_list"
	WSMsgConditionUpdate  WSMessageType = "condition_update"
)
