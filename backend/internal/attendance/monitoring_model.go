package attendance

import "time"

type AdminAttendanceState string

const (
	AdminAttendanceStateNotCheckedIn AdminAttendanceState = "NOT_CHECKED_IN"
	AdminAttendanceStateCheckedIn    AdminAttendanceState = "CHECKED_IN"
	AdminAttendanceStateCheckedOut   AdminAttendanceState = "CHECKED_OUT"
)

type AdminAttendanceSummary struct {
	Date            string `json:"date"`
	ActiveEmployees int    `json:"active_employees"`
	CheckedIn       int    `json:"checked_in"`
	CheckedOut      int    `json:"checked_out"`
	NotCheckedIn    int    `json:"not_checked_in"`
	Late            int    `json:"late"`
}

type AdminAttendanceEmployee struct {
	ID             string  `json:"id"`
	EmployeeNumber string  `json:"employee_number"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Position       *string `json:"position"`
}

type AdminAttendanceOfficeLocation struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RadiusMeters int    `json:"radius_meters"`
}

type AdminAttendanceLocationEvidence struct {
	OfficeLocationID   string  `json:"office_location_id"`
	OfficeLocationName string  `json:"office_location_name"`
	RadiusMeters       int     `json:"radius_meters"`
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	AccuracyMeters     float64 `json:"accuracy_meters"`
	DistanceMeters     float64 `json:"distance_meters"`
	InsideGeofence     bool    `json:"inside_geofence"`
}

type AdminAttendanceListItem struct {
	ID              *string                        `json:"id"`
	AttendanceDate  string                         `json:"attendance_date"`
	Employee        AdminAttendanceEmployee        `json:"employee"`
	Schedule        WorkSchedule                   `json:"schedule"`
	CheckInAt       *time.Time                     `json:"check_in_at"`
	CheckOutAt      *time.Time                     `json:"check_out_at"`
	AttendanceState AdminAttendanceState           `json:"attendance_state"`
	IsLate          bool                           `json:"is_late"`
	OfficeLocation  *AdminAttendanceOfficeLocation `json:"office_location,omitempty"`
}

type AdminAttendanceDetail struct {
	ID               string                           `json:"id"`
	AttendanceDate   string                           `json:"attendance_date"`
	Employee         AdminAttendanceEmployee          `json:"employee"`
	Schedule         WorkSchedule                     `json:"schedule"`
	CheckInAt        time.Time                        `json:"check_in_at"`
	CheckOutAt       *time.Time                       `json:"check_out_at"`
	AttendanceState  AdminAttendanceState             `json:"attendance_state"`
	IsLate           bool                             `json:"is_late"`
	CheckInLocation  *AdminAttendanceLocationEvidence `json:"check_in_location"`
	CheckOutLocation *AdminAttendanceLocationEvidence `json:"check_out_location"`
}

type AdminAttendanceList struct {
	Items      []AdminAttendanceListItem `json:"items"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalItems int                       `json:"total_items"`
	TotalPages int                       `json:"total_pages"`
}

type AdminAttendanceListFilter struct {
	DateFrom        string
	DateTo          string
	EmployeeID      string
	Search          string
	AttendanceState AdminAttendanceState
	IsLate          *bool
	Page            int
	PageSize        int
}

type adminAttendanceQuery struct {
	DateFrom        time.Time
	DateTo          time.Time
	EmployeeID      string
	Search          string
	AttendanceState AdminAttendanceState
	IsLate          *bool
	Page            int
	PageSize        int
	Timezone        string
}

type adminAttendanceSummaryRow struct {
	ActiveEmployees int
	CheckedIn       int
	CheckedOut      int
	NotCheckedIn    int
	Late            int
}

type adminAttendanceRow struct {
	ID               *string
	AttendanceDate   time.Time
	Employee         AdminAttendanceEmployee
	Schedule         WorkSchedule
	CheckInAt        *time.Time
	CheckOutAt       *time.Time
	AttendanceState  AdminAttendanceState
	IsLate           bool
	OfficeLocation   *AdminAttendanceOfficeLocation
	CheckInLocation  *AdminAttendanceLocationEvidence
	CheckOutLocation *AdminAttendanceLocationEvidence
}
