package attendance

import (
	"time"

	"r3-ti-faceattend/backend/internal/user"
)

type AttendanceState string

const (
	StateNotCheckedIn AttendanceState = "NOT_CHECKED_IN"
	StateCheckedIn    AttendanceState = "CHECKED_IN"
	StateCompleted    AttendanceState = "COMPLETED"
)

type WorkSchedule struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	GraceMinutes int       `json:"grace_minutes"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type AttendanceRecord struct {
	ID               string
	UserID           string
	ScheduleID       string
	AttendanceDate   time.Time
	CheckInAt        time.Time
	CheckOutAt       *time.Time
	CheckInLocation  *AttendanceLocationEvidence
	CheckOutLocation *AttendanceLocationEvidence
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type DailyStatus struct {
	AttendanceDate   string                      `json:"attendance_date"`
	Schedule         WorkSchedule                `json:"schedule"`
	CheckInAt        *time.Time                  `json:"check_in_at"`
	CheckOutAt       *time.Time                  `json:"check_out_at"`
	CheckInLocation  *AttendanceLocationEvidence `json:"check_in_location"`
	CheckOutLocation *AttendanceLocationEvidence `json:"check_out_location"`
	State            AttendanceState             `json:"state"`
	CanCheckIn       bool                        `json:"can_check_in"`
	CanCheckOut      bool                        `json:"can_check_out"`
}

type HistoryItem struct {
	ID               string                      `json:"id"`
	AttendanceDate   string                      `json:"attendance_date"`
	Schedule         WorkSchedule                `json:"schedule"`
	CheckInAt        time.Time                   `json:"check_in_at"`
	CheckOutAt       *time.Time                  `json:"check_out_at"`
	CheckInLocation  *AttendanceLocationEvidence `json:"check_in_location"`
	CheckOutLocation *AttendanceLocationEvidence `json:"check_out_location"`
	State            AttendanceState             `json:"state"`
}

type HistoryList struct {
	Items      []HistoryItem `json:"items"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalItems int           `json:"total_items"`
	TotalPages int           `json:"total_pages"`
}

type TodayData struct {
	User     user.User
	Schedule WorkSchedule
	Record   *AttendanceRecord
}

type HistoryRow struct {
	Record   AttendanceRecord
	Schedule WorkSchedule
}

type HistoryFilter struct {
	Page     int
	PageSize int
}

type AttendanceLocationRequest struct {
	Latitude       float64
	Longitude      float64
	AccuracyMeters float64
}

type AttendanceLocationTarget struct {
	OfficeLocationID   string
	OfficeLocationName string
	Latitude           float64
	Longitude          float64
	RadiusMeters       int
	IsActive           bool
}

type AttendanceLocationEvidence struct {
	OfficeLocationID   string  `json:"office_location_id"`
	OfficeLocationName string  `json:"office_location_name"`
	Latitude           float64 `json:"-"`
	Longitude          float64 `json:"-"`
	AccuracyMeters     float64 `json:"accuracy_meters"`
	DistanceMeters     float64 `json:"distance_meters"`
	InsideGeofence     bool    `json:"inside_geofence"`
}
