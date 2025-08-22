package domainmodels

import "time"

type GridDB struct {
	ID    int64     `json:"id" db:"id"`     // PK
	DimX  int64     `json:"dimX" db:"dim_x"`
	DimY  int64     `json:"dimY" db:"dim_y"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CellDB struct {
	ID        int64     `db:"id"`
	GridID    int64     `db:"grid_id"`
	Xpos      int64     `db:"xpos"`
	Ypos      int64     `db:"ypos"`
	CellType  string    `db:"cell_type"`         // normal, refuel, depot, blocked
	RefuelAmount *float64 `db:"refuel_amount"`   // nullable
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}


type RoadSegmentDB struct {
	ID        int64     `json:"id" db:"id"` 
	StartX    int64     `json:"start_x" db:"start_x"`
	StartY    int64     `json:"start_y" db:"start_y"`
	EndX      int64     `json:"end_x" db:"end_x"`
	EndY      int64     `json:"end_y" db:"end_y"`


	//these are optional so can be null or ommited 
	SpeedLimit *int64   `json:"speed_limit,omitempty" db:"speed_limit"`
	Capacity   *int64   `json:"capacity,omitempty" db:"capacity"`
	Weather    []string `json:"weather_conditions,omitempty" db:"weather_conditions"` 
	IsOpen     bool     `json:"is_open" db:"is_open"`

	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CellRoadDB struct {
	ID            int64     `json:"id" db:"id"` 
	CellID        int64     `json:"cell_id" db:"cell_id"` // FK 
	RoadSegmentID int64     `json:"road_segment_id" db:"road_segment_id"` // FK 
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
