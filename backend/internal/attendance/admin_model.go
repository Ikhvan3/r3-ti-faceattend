package attendance

import (
	"time"

	"r3-ti-faceattend/backend/internal/user"
)

type ScheduleStatus string

const (
	ScheduleStatusActive   ScheduleStatus = "ACTIVE"
	ScheduleStatusInactive ScheduleStatus = "INACTIVE"
)

type AssignmentStatus string

const (
	AssignmentStatusCurrent  AssignmentStatus = "CURRENT"
	AssignmentStatusUpcoming AssignmentStatus = "UPCOMING"
	AssignmentStatusEnded    AssignmentStatus = "ENDED"
)

type WorkScheduleListFilter struct {
	Page     int
	PageSize int
	Search   string
	Status   ScheduleStatus
}

type WorkScheduleList struct {
	Items      []WorkSchedule `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalItems int            `json:"total_items"`
	TotalPages int            `json:"total_pages"`
}

type WorkScheduleInput struct {
	Name         string
	StartTime    string
	EndTime      string
	GraceMinutes int
}

type AssignmentListFilter struct {
	Page       int
	PageSize   int
	Search     string
	UserID     string
	ScheduleID string
	Status     AssignmentStatus
}

type ScheduleAssignment struct {
	ID            string               `json:"id"`
	User          user.EmployeeProfile `json:"user"`
	Schedule      WorkSchedule         `json:"schedule"`
	EffectiveFrom string               `json:"effective_from"`
	EffectiveTo   *string              `json:"effective_to"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type ScheduleAssignmentRecord struct {
	ID            string
	User          user.User
	Schedule      WorkSchedule
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ScheduleAssignmentList struct {
	Items      []ScheduleAssignment `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalItems int                  `json:"total_items"`
	TotalPages int                  `json:"total_pages"`
}

type AssignmentCreateInput struct {
	UserID        string
	ScheduleID    string
	EffectiveFrom string
	EffectiveTo   *string
}
