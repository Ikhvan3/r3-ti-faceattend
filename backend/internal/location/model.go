package location

import (
	"time"

	"r3-ti-faceattend/backend/internal/user"
)

type OfficeStatus string

const (
	OfficeStatusActive   OfficeStatus = "ACTIVE"
	OfficeStatusInactive OfficeStatus = "INACTIVE"
)

type AssignmentStatus string

const (
	AssignmentStatusCurrent  AssignmentStatus = "CURRENT"
	AssignmentStatusUpcoming AssignmentStatus = "UPCOMING"
	AssignmentStatusEnded    AssignmentStatus = "ENDED"
)

type OfficeLocation struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Address      *string   `json:"address"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type OfficeLocationInput struct {
	Name         string
	Address      *string
	Latitude     float64
	Longitude    float64
	RadiusMeters int
}

type OfficeLocationListFilter struct {
	Page     int
	PageSize int
	Search   string
	Status   OfficeStatus
}

type OfficeLocationList struct {
	Items      []OfficeLocation `json:"items"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalItems int              `json:"total_items"`
	TotalPages int              `json:"total_pages"`
}

type LocationAssignmentInput struct {
	UserID           string
	OfficeLocationID string
	EffectiveFrom    string
	EffectiveTo      *string
}

type LocationAssignmentListFilter struct {
	Page             int
	PageSize         int
	Search           string
	UserID           string
	OfficeLocationID string
	Status           AssignmentStatus
}

type LocationAssignmentRecord struct {
	ID            string
	User          user.User
	Office        OfficeLocation
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type LocationAssignment struct {
	ID            string               `json:"id"`
	User          user.EmployeeProfile `json:"user"`
	Office        OfficeLocation       `json:"office_location"`
	EffectiveFrom string               `json:"effective_from"`
	EffectiveTo   *string              `json:"effective_to"`
	Status        AssignmentStatus     `json:"status"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type LocationAssignmentList struct {
	Items      []LocationAssignment `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalItems int                  `json:"total_items"`
	TotalPages int                  `json:"total_pages"`
}

type LocationRequirement struct {
	Assignment LocationAssignment `json:"assignment"`
	Office     OfficeLocation     `json:"office_location"`
}
